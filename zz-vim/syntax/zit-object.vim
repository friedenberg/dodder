
if exists("b:current_syntax")
  finish
endif

if $BIN_ZIT == ""
  let $BIN_ZIT = "zit"
endif

let zettel = expand("%")

let g:markdown_syntax_conceal = 0

if zettel != ""
  let cmdFormat = "$BIN_ZIT show -quiet -format type.vim-syntax-type " . zettel
  let zettelTypSyntax = trim(system(cmdFormat))

  if v:shell_error
    echom "Error getting vim syntax type: " . zettelTypSyntax
    let zettelTypSyntax = "markdown"
  elseif zettelTypSyntax == ""
    echom "Zettel Type has no vim syntax set"
    let zettelTypSyntax = "markdown"
  endif

  let zit_syntax_path = $HOME."/.local/share/zit/vim/syntax/".zettelTypSyntax.".vim"
  let vim_syntax_path = $VIMRUNTIME."/syntax/" . zettelTypSyntax . ".vim"

  if filereadable(zit_syntax_path)
    execute "syntax include @akte" zit_syntax_path
  elseif filereadable(vim_syntax_path)
    execute "syntax include @akte" vim_syntax_path
  else
    echom "could not find syntax file for ".zettelTypSyntax
  endif
endif

syn region zitAkte start=// end=// contains=@akte
" TODO set comment strings for body

let m = expand("<sfile>:h") . "/zit-metadata.vim"
exec "source " . m

let b:current_syntax = 'zit-object'
