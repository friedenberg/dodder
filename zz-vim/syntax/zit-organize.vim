
" if exists("b:current_syntax")
"   finish
" endif

let m = expand("<sfile>:h") . "/zit-metadata.vim"
exec "source " . m

" syn match zitSkuTagComponent '\w\+' contained
" syn region zitSkuTag start=/\w/ end=' ' contained contains=@NoSpell,zitSkuTagComponent

syn match zitSkuFieldValue /.\+/ contained
syn match zitSkuFieldEscape /\\./ contained
syn region zitSkuField start=/"/ skip=/\\./ end=/"/ keepend contained 
      \ contains=zitSkuFieldValue,zitSkuFieldEscape

syn match zitSkuTypeComponent '\w\+' contained
syn region zitSkuType start=/!/ms=e+1 end=' ' contained contains=@NoSpell,zitSkuTypeComponent

" don't include the newline because this is within a region
syn match zitSkuDescription /\v.*/ contained

syn match zitSkuObjectIdComponent '\w\+' contained
syn region zitSkuObjectId start='\[\s*'ms=e+1 end=' ' contained 
      \ contains=zitSkuObjectIdComponent

syn region zitSkuMetadataRegion start='\[' end='\]' keepend
      \ contains=zitSkuObjectId,zitSkuField,zitSkuType
      \ nextgroup=zitSkuDescription

syn match zitTag /\v[^#,]+/ contained contains=@NoSpell
syn match zitTagPrefix /\v#+/ contained
syn region zitTagRegion start=/\v^\s*#+ / end=/$/
      \ contains=zitTag,zitTagPrefix

highlight default link zitTag Title
highlight default link zitSkuObjectIdComponent Identifier
highlight default link zitSkuTypeComponent Type
highlight default link zitSkuFieldValue Constant
highlight default link zitSkuSyntax Normal
highlight default link zitSkuFieldEscape SpecialChar
highlight default link zitSkuDescription String

" debug
" highlight default link zitSkuMetadataRegion Underlined

let b:current_syntax = 'zit-organize'
