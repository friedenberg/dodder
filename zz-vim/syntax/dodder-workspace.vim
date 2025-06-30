
if exists("b:current_syntax")
  finish
endif

if $BIN_DODDER == ""
  let $BIN_DODDER = "dodder"
endif

let zettel = expand("%")

let g:markdown_syntax_conceal = 0

if zettel != ""
  " TODO add dodder command for loading type from triple-hyphen-io docs
  " let cmdFormat = "$BIN_DODDER show -quiet -format type.vim-syntax-type " . zettel
  let zettelTypSyntax = "toml"

  " if v:shell_error
  "   echom "Error getting vim syntax type: " . zettelTypSyntax
  "   let zettelTypSyntax = "markdown"
  " elseif zettelTypSyntax == ""
  "   echom "Zettel Type has no vim syntax set"
  "   let zettelTypSyntax = "markdown"
  " endif

  let dodder_syntax_path = $HOME."/.local/share/dodder/vim/syntax/".zettelTypSyntax.".vim"
  let vim_syntax_path = $VIMRUNTIME."/syntax/" . zettelTypSyntax . ".vim"

  if filereadable(dodder_syntax_path)
    execute "syntax include @akte" dodder_syntax_path
  elseif filereadable(vim_syntax_path)
    execute "syntax include @akte" vim_syntax_path
  else
    echom "could not find syntax file for ".zettelTypSyntax
  endif
endif

syn region dodderAkte start=// end=// contains=@akte

let m = expand("<sfile>:h") . "/dodder-metadata.vim"
exec "source " . m

let b:current_syntax = 'dodder-workspace'
