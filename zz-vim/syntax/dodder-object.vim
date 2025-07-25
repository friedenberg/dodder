
if exists("b:current_syntax")
  finish
endif

if $BIN_DODDER == ""
  let $BIN_DODDER = "der"
endif

let object = expand("%")

let g:markdown_syntax_conceal = 0

if object != ""
  let cmdFormat = "$BIN_DODDER show -quiet -format type.vim-syntax-type ".object
  let objectTypeSyntax = trim(system(cmdFormat))

  if v:shell_error
    echom "Error getting vim syntax type: ".objectTypeSyntax
    echom "Error: ".objectTypeSyntax
    " TODO use default objectTypeSyntax
    let objectTypeSyntax = "pandoc"
  elseif objectTypeSyntax == ""
    echom "Object Type has no vim syntax set"
    let objectTypeSyntax = "markdown"
  endif

  let dodder_syntax_path = $HOME."/.local/share/dodder/vim/syntax/".objectTypeSyntax.".vim"
  let vim_syntax_path = $VIMRUNTIME."/syntax/" . objectTypeSyntax . ".vim"

  if filereadable(dodder_syntax_path)
    execute "syntax include @akte" dodder_syntax_path
  elseif filereadable(vim_syntax_path)
    execute "syntax include @akte" vim_syntax_path
  else
    echom "could not find syntax file for ".objectTypeSyntax
  endif
endif

syn region dodderAkte start=// end=// contains=@akte
" TODO set comment strings for body

let m = expand("<sfile>:h") . "/dodder-metadata.vim"
exec "source " . m

let b:current_syntax = 'dodder-object'
