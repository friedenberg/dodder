
let &l:commentstring = "# %s"

" let &l:equalprg = "zit format-zettel %"
" let &l:comments = "fb:*,fb:-,fb:+,n:>"
" let &l:commentstring = "<!--%s-->"

" function! Gf()
"   let l:f = expand("<cfile>") . ".md"

"   if !filereadable(l:f)
"     echo system("zit checkout -mode both " . expand("<cfile>"))
"   endif

"   execute 'tabedit' l:f
"   " try
"   "   " exec "normal! \<c-w>gf"
"   " catch /E447/
"   " endtry
" endfunction

" noremap gf :call Gf()<CR>

" " TODO support external akte
" function! ZitTypActionMenu()
"   let l:items = systemlist("zit show -format typ-action-names " . expand("%:r"))

"   func! ZitTypActionMenuItemPicked(id, result) closure
"     if a:result == -1
"       return
"     endif

"     let l:val = substitute(l:items[a:result-1], '\t.*$', '', '')
"     execute("!zit exec-action -action " . l:val .  " " . expand("%:r"))
"   endfunc

"   if len(l:items) == 0
"     echom "No Zettel-Typ-specific actions available"
"     return
"   endif

"   call popup_menu(
"         \ items,
"         \ #{ title: "Run a Zettel-Typ-Specific Action", 
"         \ callback: 'ZitTypActionMenuItemPicked', 
"         \ line: 25, col: 40,
"         \ highlight: 'Question', border: [], close: 'click',  padding: [1,1,0,1]} )
" endfunction

" let maplocalleader = "-"

" nnoremap <localleader>z :call ZitTypActionMenu()<cr>
