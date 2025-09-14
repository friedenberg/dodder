package remote_http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"github.com/gorilla/mux"
)

// TODO use context cancellation for http errors

type Server struct {
	EnvLocal  env_local.Env
	Repo      *local_working_copy.Repo
	blobCache serverBlobCache

	GetCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error)
}

func (server *Server) init() (err error) {
	server.blobCache.localBlobStore = server.Repo.GetEnvRepo().GetDefaultBlobStore()
	server.blobCache.ui = server.Repo.GetEnv().GetUI()
	return
}

// TODO switch to not return error
func (server *Server) InitializeListener(
	network, address string,
) (listener net.Listener, err error) {
	var config net.ListenConfig

	switch network {
	case "unix":
		if listener, err = server.InitializeUnixSocket(config, address); err != nil {
			err = errors.Wrap(err)
			return
		}

	case "tcp":
		if listener, err = config.Listen(
			server.Repo.GetEnv(),
			network,
			address,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		addr := listener.Addr().(*net.TCPAddr)

		server.EnvLocal.GetOut().Printf(
			"starting HTTP server on port: %q",
			strconv.Itoa(addr.Port),
		)

	default:
		if listener, err = config.Listen(
			server.Repo.GetEnv(),
			network,
			address,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (server *Server) InitializeUnixSocket(
	config net.ListenConfig,
	path string,
) (sock repo.UnixSocket, err error) {
	sock.Path = path

	if sock.Path == "" {
		dir := server.EnvLocal.GetXDG().State

		if err = os.MkdirAll(dir, 0o700); err != nil {
			err = errors.Wrap(err)
			return
		}

		sock.Path = fmt.Sprintf("%s/%d.sock", dir, os.Getpid())
	}

	ui.Log().Printf("starting unix domain server on socket: %q", sock.Path)

	if sock.Listener, err = config.Listen(
		server.Repo.GetEnv(),
		"unix",
		sock.Path,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type HTTPPort struct {
	net.Listener
	Port int
}

func (server *Server) InitializeHTTP(
	config net.ListenConfig,
	port int,
) (httpPort HTTPPort, err error) {
	if httpPort.Listener, err = config.Listen(
		server.Repo.GetEnv(),
		"tcp",
		fmt.Sprintf(":%d", port),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	addr := httpPort.Addr().(*net.TCPAddr)

	ui.Log().Printf("starting HTTP server on port: %q", strconv.Itoa(addr.Port))

	return
}

func (server *Server) makeRouter(
	makeHandler func(handler funcHandler) http.HandlerFunc,
) http.Handler {
	// TODO add errors/context middlerware for capturing errors and panics
	router := mux.NewRouter().UseEncodedPath()

	router.HandleFunc(
		"/config-immutable",
		makeHandler(server.handleGetConfigImmutable),
	).
		Methods(
			"GET",
		)

	{
		router.HandleFunc(
			"/blobs/{blob_id}",
			makeHandler(server.handleBlobsHeadOrGet),
		).
			Methods(
				"HEAD",
				"GET",
			)

		router.HandleFunc(
			"/blobs/{blob_id}",
			makeHandler(server.handleBlobsPost),
		).
			Methods(
				"POST",
			)

		router.HandleFunc("/blobs", makeHandler(server.handleBlobsPost)).
			Methods("POST")
	}

	router.HandleFunc(
		"/query/{list_type}/{query}",
		makeHandler(server.handleGetQuery),
	).
		Methods(
			"GET",
		)

	router.HandleFunc("/mcp", makeHandler(server.handleMCP)).
		Methods("POST")

	{
		router.HandleFunc(
			"/inventory_lists",
			makeHandler(server.handleGetInventoryList),
		).
			Methods(
				"GET",
			)

		router.HandleFunc(
			"/inventory_lists",
			makeHandler(server.handlePostInventoryList),
		).
			Methods(
				"POST",
			)

		router.HandleFunc(
			"/inventory_lists/{list_type}/{list_object}",
			makeHandler(server.handlePostInventoryList),
		).
			Methods(
				"POST",
			)
	}

	if server.Repo.GetEnv().GetCLIConfig().Verbose {
		router.Use(server.loggerMiddleware)
	}

	router.Use(server.panicHandlingMiddleware)
	router.Use(server.sigMiddleware)

	return router
}

func (server *Server) addSignatureIfNecessary(
	nonceString string,
	header http.Header,
) (err error) {
	if nonceString == "" {
		err = errors.Errorf("nonce empty or not provided")
		return
	}

	var nonce markl.Id

	if err = nonce.Set(
		nonceString,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	header.Set(
		headerRepoPublicKey,
		server.Repo.GetImmutableConfigPublic().GetPublicKey().String(),
	)

	privateKey := server.Repo.GetImmutableConfigPrivate().Blob.GetPrivateKey()

	var sig markl.Id

	if err = markl.Sign(
		privateKey,
		nonce,
		markl.PurposeRequestAuthResponseV1,
		markl.TypeIdEd25519Sig,
		&sig,
	); err != nil {
		server.EnvLocal.Cancel(err)
		return
	}

	header.Set(headerChallengeResponse, sig.String())

	return
}

func (server *Server) sigMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(responseWriter http.ResponseWriter, request *http.Request) {
			if err := server.addSignatureIfNecessary(
				request.Header.Get(headerChallengeNonce),
				responseWriter.Header(),
			); err != nil {
				http.Error(responseWriter, err.Error(), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(responseWriter, request)
		},
	)
}

func (server *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(responseWriter http.ResponseWriter, request *http.Request) {
			ui.Log().Printf(
				"serving request: %s %s",
				request.Method,
				request.URL.Path,
			)
			next.ServeHTTP(responseWriter, request)
			ui.Log().Printf(
				"done serving request: %s %s",
				request.Method,
				request.URL.Path,
			)
		},
	)
}

// TODO consider removing in place of context
func (server *Server) panicHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(responseWriter http.ResponseWriter, request *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					ui.Log().Print("request handler panicked", request.URL)

					switch err := r.(type) {
					default:
						panic(err)

					case error:
						http.Error(
							responseWriter,
							fmt.Sprintf("%s: %s", err, debug.Stack()),
							http.StatusInternalServerError,
						)
					}
				}
			}()

			next.ServeHTTP(responseWriter, request)
		},
	)
}

// TODO remove error return and use context
func (server *Server) Serve(listener net.Listener) (err error) {
	if err = server.init(); err != nil {
		err = errors.Wrap(err)
		return
	}

	httpServer := http.Server{
		Handler: server.makeRouter(server.makeHandler),
	}

	if server.GetCertificate != nil {
		httpServer.TLSConfig = &tls.Config{
			GetCertificate: server.GetCertificate,
		}
	}

	go func() {
		<-server.Repo.GetEnv().Done()
		ui.Log().Print("shutting down")

		ctx, cancel := context.WithTimeoutCause(
			context.Background(),
			1e9, // 1 second
			errors.ErrorWithStackf("shut down timeout"),
		)

		defer cancel()

		httpServer.Shutdown(ctx)
	}()

	if err = httpServer.Serve(listener); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	ui.Log().Print("shutdown complete")

	return
}

func (server *Server) ServeStdio() {
	listener := MakeStdioListener()

	if err := server.Serve(listener); err != nil {
		server.EnvLocal.Cancel(err)
		return
	}
}

type funcHandler func(Request) Response

type handlerWrapper funcHandler

func (server *Server) makeHandler(
	handler funcHandler,
) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, req *http.Request) {
		request := Request{
			ctx:        errors.MakeContext(server.EnvLocal),
			request:    req,
			MethodPath: MethodPath{Method: req.Method, Path: req.URL.Path},
			Headers:    req.Header,
			Body:       req.Body,
		}

		var progressWriter env_ui.ProgressWriter
		var response Response

		if err := errors.RunContextWithPrintTicker(
			request.ctx,
			func(ctx interfaces.Context) {
				response = handler(request)
			},
			func(time time.Time) {
				ui.Log().Printf(
					"Still serving request (%s): %q",
					time,
					req.URL,
					progressWriter.GetWrittenHumanString(),
				)
			},
			3*time.Second,
		); err != nil {
			_, frames := request.ctx.CauseWithStackFrames()

			err = stack_frame.MakeErrorTreeOrErr(err, frames...)

			response.Error(err)
		}

		if err := errors.RunContextWithPrintTicker(
			request.ctx,
			func(ctx interfaces.Context) {
				header := responseWriter.Header()

				for key, values := range response.Headers() {
					for _, value := range values {
						header.Add(key, value)
					}
				}

				if response.StatusCode == 0 {
					response.StatusCode = http.StatusOK
				}

				responseWriter.WriteHeader(response.StatusCode)

				if response.Body == nil {
					return
				}

				if _, err := io.Copy(
					io.MultiWriter(responseWriter, &progressWriter),
					response.Body,
				); err != nil {
					if errors.IsEOF(err) {
						err = nil
					} else if errors.IsAny(
						err,
						errors.MakeIsErrno(
							syscall.ECONNRESET,
							syscall.EPIPE,
						),
						errors.IsNetTimeout,
					) {
						ui.Err().Print(errors.Unwrap(err).Error(), req.URL)
						err = nil
					} else {
						ctx.Cancel(err)
					}
				}
			},
			func(time time.Time) {
				ui.Log().Printf(
					"Still serving request (%s): %q (%s bytes written)",
					time,
					req.URL,
					progressWriter.GetWrittenHumanString(),
				)
			},
			3*time.Second,
		); err != nil {
			_, frames := request.ctx.CauseWithStackFrames()

			err = stack_frame.MakeErrorTreeOrErr(err, frames...)

			http.Error(
				responseWriter,
				err.Error(),
				http.StatusInternalServerError,
			)
		}
	}
}

func (server *Server) handleBlobsHeadOrGet(
	request Request,
) (response Response) {
	// TODO rename to blob id
	blobIdString := request.Vars()["blob_id"]

	if blobIdString == "" {
		response.ErrorWithStatus(
			http.StatusBadRequest,
			errors.ErrorWithStackf("empty blob id"),
		)
		return
	}

	var blobId markl.Id

	{
		var err error

		if err = blobId.Set(
			blobIdString,
		); err != nil {
			response.ErrorWithStatus(http.StatusBadRequest, err)
			return
		}
	}

	ui.Log().Printf("blob requested: %q", blobId)

	if request.Method == "HEAD" {
		if server.Repo.GetBlobStore().HasBlob(blobId) {
			response.StatusCode = http.StatusNoContent
		} else {
			response.StatusCode = http.StatusNotFound
		}
	} else {
		var rc interfaces.BlobReader

		{
			var err error

			if rc, err = server.Repo.GetBlobStore().MakeBlobReader(blobId); err != nil {
				if env_dir.IsErrBlobMissing(err) {
					response.StatusCode = http.StatusNotFound
				} else {
					response.Error(err)
				}

				return
			}
		}

		response.Body = rc
	}

	return
}

func (server *Server) handleBlobsPost(request Request) (response Response) {
	blobId := request.Vars()["blob_id"]
	var copyResult blob_stores.CopyResult

	if blobId == "" {
		var err error

		if copyResult, err = server.copyBlob(request.Body, nil); err != nil {
			response.Error(err)
			return
		}

		response.StatusCode = http.StatusCreated
		response.Body = io.NopCloser(
			strings.NewReader(copyResult.BlobId.String()),
		)

		return
	}

	var blobDigest markl.Id

	if err := blobDigest.Set(
		blobId,
	); err != nil {
		response.Error(err)
		return
	}

	if server.Repo.GetBlobStore().HasBlob(&blobDigest) {
		response.StatusCode = http.StatusFound
		return
	}

	{
		var err error

		if copyResult, err = server.copyBlob(request.Body, &blobDigest); err != nil {
			response.Error(err)
			return
		}
	}

	response.StatusCode = http.StatusCreated

	if err := markl.AssertEqual(&blobDigest, copyResult.BlobId); err != nil {
		response.Error(err)
		return
	}

	response.StatusCode = http.StatusCreated
	response.Body = io.NopCloser(
		strings.NewReader(copyResult.BlobId.String()),
	)

	return
}

func (server *Server) copyBlob(
	reader io.ReadCloser,
	expected interfaces.MarklId,
) (copyResult blob_stores.CopyResult, err error) {
	var progressWriter env_ui.ProgressWriter
	var writeCloser interfaces.BlobWriter

	if writeCloser, err = server.Repo.GetBlobStore().MakeBlobWriter(
		nil,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	blobExpectedIdString := "blob with unknown blob id"

	if expected != nil {
		blobExpectedIdString = expected.String()
	}

	copyResult = blob_stores.CopyReaderToWriter(
		server.EnvLocal,
		writeCloser,
		reader,
		expected,
		&progressWriter,
		func(time time.Time) {
			ui.Err().Printf(
				"Copying %s... (%s written)",
				blobExpectedIdString,
				progressWriter.GetWrittenHumanString(),
			)
		},
		3*time.Second,
	)

	// ui.Debug().Print(expected, copyResult.BlobId)

	// TODO cache this
	blobCopierDelegate := sku.MakeBlobCopierDelegate(
		server.Repo.GetEnv().GetUI(),
		false,
	)

	if err = blobCopierDelegate(
		sku.BlobCopyResult{
			CopyResult: copyResult,
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (server *Server) handleGetQuery(request Request) (response Response) {
	var listTypeString string

	{
		var err error

		if listTypeString, err = url.QueryUnescape(
			request.Vars()["list_type"],
		); err != nil {
			response.Error(err)
			return
		}
	}

	var queryGroupString string

	{
		var err error

		if queryGroupString, err = url.QueryUnescape(
			request.Vars()["query"],
		); err != nil {
			response.Error(err)
			return
		}
	}

	var queryGroup *query.Query

	{
		var err error

		if queryGroup, err = server.Repo.MakeExternalQueryGroup(
			nil,
			sku.ExternalQueryOptions{},
			queryGroupString,
		); err != nil {
			response.Error(err)
			return
		}
	}

	var list *sku.ListTransacted

	{
		var err error

		if list, err = server.Repo.MakeInventoryList(queryGroup); err != nil {
			response.Error(err)
			return
		}
	}

	// TODO make this more performant by returning a proper reader
	buffer := bytes.NewBuffer(nil)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(buffer)
	defer repoolBufferedWriter()

	inventoryListCoderCloset := server.Repo.GetInventoryListCoderCloset()

	if _, err := inventoryListCoderCloset.WriteBlobToWriter(
		server.Repo,
		ids.GetOrPanic(listTypeString).Type,
		quiter.MakeSeqErrorFromSeq(list.All()),
		bufferedWriter,
	); err != nil {
		server.EnvLocal.Cancel(err)
	}

	if err := bufferedWriter.Flush(); err != nil {
		server.EnvLocal.Cancel(err)
	}

	response.Body = io.NopCloser(buffer)

	return
}

func (server *Server) handleGetInventoryList(
	request Request,
) (response Response) {
	inventoryListStore := server.Repo.GetInventoryListStore()

	// TODO make this more performant by returning a proper reader
	buffer := bytes.NewBuffer(nil)

	// TODO replace with sku.ListFormat
	boxFormat := box_format.MakeBoxTransactedArchive(
		server.Repo.GetEnv(),
		server.Repo.GetConfig().GetPrintOptions().WithPrintTai(true),
	)

	printer := string_format_writer.MakeDelim(
		"\n",
		buffer,
		string_format_writer.MakeFunc(
			func(
				writer interfaces.WriterAndStringWriter,
				object *sku.Transacted,
			) (n int64, err error) {
				return boxFormat.EncodeStringTo(object, writer)
			},
		),
	)

	iter := inventoryListStore.AllInventoryLists()

	for sk, err := range iter {
		if err != nil {
			response.Error(err)
			return
		}

		errors.ContextContinueOrPanic(server.Repo.GetEnv())

		if err = printer(sk); err != nil {
			response.Error(err)
			return
		}
	}

	response.Body = io.NopCloser(buffer)

	return
}

func (server *Server) handlePostInventoryList(
	request Request,
) (response Response) {
	listTypeString := request.Vars()["list_type"]
	listObjectString := request.Vars()["list_object"]

	if listTypeString == "" || listObjectString == "" {
		if listTypeString != "" {
			response.ErrorWithStatus(
				http.StatusBadRequest,
				errors.BadRequestf("no list type provided"),
			)
			return
		} else if listObjectString != "" {
			response.ErrorWithStatus(
				http.StatusBadRequest,
				errors.BadRequestf("no list object provided"),
			)
			return
		} else {
			return server.handlePostInventoryListNew(request)
		}
	}

	inventoryListCoderCloset := server.Repo.GetInventoryListCoderCloset()

	{
		var err error

		if listTypeString, err = url.QueryUnescape(listTypeString); err != nil {
			response.Error(err)
			return
		}
	}

	{
		var err error

		if listObjectString, err = url.QueryUnescape(listObjectString); err != nil {
			response.Error(err)
			return
		}
	}

	var object *sku.Transacted

	{
		var err error

		bufferedReader, repool := pool.GetBufferedReader(
			strings.NewReader(listObjectString),
		)
		defer repool()

		if object, err = inventoryListCoderCloset.ReadInventoryListObject(
			request.ctx,
			ids.MustType(listTypeString),
			bufferedReader,
		); err != nil {
			response.Error(
				errors.ErrorWithStackf(
					"failed to parse inventory list sku (%q): %w",
					listObjectString,
					err,
				),
			)

			return
		}

		defer sku.GetTransactedPool().Put(object)
	}

	response = server.writeInventoryListLocalWorkingCopy(
		server.Repo,
		request,
		object,
	)

	return
}

func (server *Server) handlePostInventoryListNew(
	request Request,
) (response Response) {
	response = server.writeInventoryListTypedBlobLocalWorkingCopy(
		server.Repo,
		request,
	)

	return
}

func (server *Server) handleGetConfigImmutable(
	request Request,
) (response Response) {
	configLoaded := &genesis_configs.TypedConfigPublic{
		Type: server.Repo.GetImmutableConfigPublicType(),
		Blob: server.Repo.GetImmutableConfigPublic(),
	}

	var buffer bytes.Buffer

	// TODO modify to not have to buffer
	if _, err := genesis_configs.CoderPublic.EncodeTo(configLoaded, &buffer); err != nil {
		server.EnvLocal.Cancel(err)
	}

	response.Body = io.NopCloser(&buffer)

	return
}
