
setlocal list
" TODO document
let &l:t_ut = ''
let &l:listchars = "tab:  ,trail:·,nbsp:·"
let &l:equalprg = "$BIN_ZIT format-organize -metadata-header %"

let &l:foldmethod = "expr"
let &l:foldexpr = "GetZitOrganizeFold()"

set foldtext=MyFoldText()
function MyFoldText()
  let line = getline(v:foldstart)
  let prefix = "+" . v:folddashes . " " . (v:foldend - v:foldstart) . " lines: "
  let sub = substitute(line, '/\*\|\*/\|{{{\d\=', '', 'g')
  let subTrimmed = sub[len(prefix):]
  return prefix . subTrimmed
endfunction

function! GetPreviousHeaderLineFoldLevel(lnum)
  let current = a:lnum - 1

  while current >= 0
    let v = getline(current)

    if v =~? '\v^\s*#'
      return count(v, "#")
    endif

    let current -= 1
  endwhile

  return -1
endfunction

" TODO implement against new organize syntax
function! GetZitOrganizeFold()
  let l = getline(a:lnum)

  if l =~? '\v^\s*#'
    let this_indent = count(l, "#")
    return '>' . this_indent
  else
    return -1
    " return GetPreviousHeaderLineFoldLevel(a:lnum)
  endif
endfunction

" TODO refactor into common
function! GfOrganize()
  let l:cfile = expand("<cfile>")
  let l:expanded = trim(system("$BIN_ZIT expand-hinweis " .. l:cfile))

  if !filereadable(l:expanded .. ".zettel")
    echom trim(system("$BIN_ZIT checkout -mode both " .. l:expanded))
  endif

  " let l:cmd = 'tabedit ' .. fnameescape(l:expanded .. ".zettel")
  execute 'tabedit' fnameescape(l:expanded .. ".zettel")
endfunction

" <buffer> restricts this remap to the buffer in which it was defined
" https://learnvimscriptthehardway.stevelosh.com/chapters/11.html
noremap <buffer> gf :call GfOrganize()<CR>

