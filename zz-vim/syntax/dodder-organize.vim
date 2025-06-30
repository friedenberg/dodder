
" if exists("b:current_syntax")
"   finish
" endif

let m = expand("<sfile>:h") . "/dodder-metadata.vim"
exec "source " . m

" syn match dodderSkuTagComponent '\w\+' contained
" syn region dodderSkuTag start=/\w/ end=' ' contained contains=@NoSpell,dodderSkuTagComponent

syn match dodderSkuFieldValue /.\+/ contained
syn match dodderSkuFieldEscape /\\./ contained
syn region dodderSkuField start=/"/ skip=/\\./ end=/"/ keepend contained 
      \ contains=dodderSkuFieldValue,dodderSkuFieldEscape

syn match dodderSkuTypeComponent '\w\+' contained
syn region dodderSkuType start=/!/ms=e+1 end=' ' contained contains=@NoSpell,dodderSkuTypeComponent

" don't include the newline because this is within a region
syn match dodderSkuDescription /\v.*/ contained

syn match dodderSkuObjectIdComponent '\w\+' contained
syn region dodderSkuObjectId start='\[\s*'ms=e+1 end=' ' contained 
      \ contains=dodderSkuObjectIdComponent

syn region dodderSkuMetadataRegion start='\[' end='\]' keepend
      \ contains=dodderSkuObjectId,dodderSkuField,dodderSkuType
      \ nextgroup=dodderSkuDescription

syn match dodderTag /\v[^#,]+/ contained contains=@NoSpell
syn match dodderTagPrefix /\v#+/ contained
syn region dodderTagRegion start=/\v^\s*#+ / end=/$/
      \ contains=dodderTag,dodderTagPrefix

highlight default link dodderTag Title
highlight default link dodderSkuObjectIdComponent Identifier
highlight default link dodderSkuTypeComponent Type
highlight default link dodderSkuFieldValue Constant
highlight default link dodderSkuSyntax Normal
highlight default link dodderSkuFieldEscape SpecialChar
highlight default link dodderSkuDescription String

" debug
" highlight default link dodderSkuMetadataRegion Underlined

let b:current_syntax = 'dodder-organize'
