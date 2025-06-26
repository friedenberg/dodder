
let &l:equalprg = "$BIN_ZIT format-object -mode both %"
" TODO-P3 use https://github.com/suy/vim-context-commentstring
let &l:comments = "fb:*,fb:-,fb:+,n:>"
let &l:commentstring = "<!--%s-->"

function! GfZettel()
  let l:h = expand("<cfile>")
  let l:expanded = trim(system("$BIN_ZIT expand-zettel-id " . l:h))
  let l:f = l:expanded . ".zettel"

  if !filereadable(l:f)
    echo system("$BIN_ZIT checkout -mode both " . l:expanded)
  endif

  let l:cmd = 'tabedit ' . l:f
  execute l:cmd
  " try
  "   " exec "normal! \<c-w>gf"
  " catch /E447/
  " endtry
endfunction

noremap <buffer> gf :call GfZettel()<CR>

" TODO support external blob
function! ZitAction()
  let [l:items, l:processedItems] = ZitGetActionNames()

  func! ZitActionItemPicked(id, result) closure
    if a:result == -1
      return
    endif

    let l:val = substitute(l:items[a:result-1], '\t.*$', '', '')
    execute("!$BIN_ZIT exec-action -action " . l:val .  " " . GetObjectId())
  endfunc

  if len(l:items) == 0
    echom "No Zettel-Typ-specific actions available"
    return
  endif

  call popup#menu(
        \ items,
        \ #{ title: "Run a Zettel-Typ-Specific Action", 
        \ callback: 'ZitActionItemPicked', 
        \ line: 25, col: 40,
        \ highlight: 'Question', border: [], close: 'click',  padding: [1,1,0,1]} )
endfunction

function! ZitMakeUTIGroupCommand(uti_group, cmd_args_unprocessed_list)
  let l:cmd_args_list = []

  let i = 0

  while i < len(a:cmd_args_unprocessed_list)
    let l:uti = a:cmd_args_unprocessed_list[i]
    let l:formatter = a:cmd_args_unprocessed_list[i+1]
    call add(l:cmd_args_list, "-i")
    call add(l:cmd_args_list, l:uti)
    let l:cmd_sub_args = [
          \ "$BIN_ZIT", "format-object", "-mode blob",
          \ "-uti-group", a:uti_group,
          \ GetObjectId(),
          \ l:uti,
          \ "2>/dev/null",
          \ ]

    call add(l:cmd_args_list, "<(" . join(l:cmd_sub_args, " ") . ")")

    let i += 2
  endwhile

  return l:cmd_args_list
endfunction

function! SplitListOnSpaceAndReturnBoth(rawItems)
  let l:processedItems = []
  let l:items = []

  for i in a:rawItems
    let l:groupName = substitute(i, '\s.*$', '', '')
    let l:group = i[len(l:groupName) +1:]
    call add(l:items, l:groupName)
    call add(l:processedItems, l:group)
  endfor

  return [l:items, l:processedItems]
endfunction

function! GetObjectId()
  return expand("%")
endfunction

function! ZitGetUTIGroups()
  let l:rawItems = sort(systemlist("$BIN_ZIT show -format type.formatter-uti-groups " . GetObjectId()))
  return SplitListOnSpaceAndReturnBoth(l:rawItems)
endfunction

function! ZitGetActionNames()
  let l:rawItems = sort(systemlist("$BIN_ZIT show -format type.action-names " . GetObjectId()))
  return SplitListOnSpaceAndReturnBoth(l:rawItems)
endfunction

function! ZitGetFormats()
  let l:rawItems =  sort(systemlist("$BIN_ZIT show -format type.formatters " . GetObjectId()))
  return SplitListOnSpaceAndReturnBoth(l:rawItems)
endfunction

function! ZitPreview()
  let [l:formatIds, l:fileExtensions] = ZitGetFormats()

  func! ZitPreviewMenuItemPicked(id, result) closure
    if a:result == -1
      return
    endif

    " let l:format = substitute(l:formatIds[a:result-1], '\t.*$', '', '')
    let l:format = l:formatIds[a:result-1]
    let l:fileExtension = l:fileExtensions[a:result-1]
    let l:objectId = GetObjectId()

    let l:tempfile = tempname() .. "." .. l:fileExtension

    let l:cmd_args_list = [
          \ "zit format-object -mode blob",
          \ l:objectId,
          \ l:format,
          \ ">",
          \ l:tempfile,
          \ ]

    let l:cmd_args =  join(l:cmd_args_list, " ")
    call system(l:cmd_args)

    let l:cmd_preview = "qlmanage -p "..l:tempfile..">/dev/null 2>&1 &"
    call system(l:cmd_preview)
  endfunc

  if len(l:formatIds) == 1
    call ZitPreviewMenuItemPicked("", 1)
    return
  endif

  if len(l:formatIds) == 0
    echom "No Zettel-Typ-specific actions available"
    return
  endif

  call popup#menu(
        \ formatIds,
        \ #{ title: "Preview format", 
        \ callback: 'ZitPreviewMenuItemPicked', 
        \ line: 25, col: 40,
        \ highlight: 'Question', border: [], close: 'click',  padding: [1,1,0,1]} )
endfunction

function! ZitCopy()
  let [l:items, l:processedItems] = ZitGetUTIGroups()

  func! ZitCopyMenuItemPicked(id, result) closure
    if a:result == -1
      return
    endif

    let l:uti_group = l:items[a:result-1]
    let l:val = substitute(l:processedItems[a:result-1], '\t.*$', '', '')
    let l:cmd_args_list = ZitMakeUTIGroupCommand(l:uti_group, split(l:val))

    execute("!tacky copy " . join(l:cmd_args_list, " "))
  endfunc

  if len(l:processedItems) == 1
    call ZitCopyMenuItemPicked("", 1)
    return
  endif

  if len(l:items) == 0
    echom "No Zettel-Typ-specific actions available"
    return
  endif

  call popup#menu(
        \ items,
        \ #{ title: "Copy format", 
        \ callback: 'ZitCopyMenuItemPicked', 
        \ line: 25, col: 40,
        \ highlight: 'Question', border: [], close: 'click',  padding: [1,1,0,1]} )
endfunction

let maplocalleader = "-"

nnoremap <localleader>z :call ZitAction()<cr>
nnoremap <localleader>c :call ZitCopy()<cr>
nnoremap <localleader>p :call ZitPreview()<cr>
