
" if exists("b:current_syntax")
"   finish
" endif

syn region dodderMetadatei start=/\v%^---$/ end=/\v^---$/ 
      \ contains=dodderMetadateiBezeichnungRegion,dodderMetadateiEtikettRegion,dodderMetadateiAkteRegion,dodderMetadateiCommentRegion
      \ nextgroup=dodderAkte

syn match dodderMetadateiBezeichnung /\v[^\n]+/ contained
syn match dodderMetadateiBezeichnungPrefix /\v^# / contained nextgroup=dodderMetadateiBezeichnung
syn region dodderMetadateiBezeichnungRegion start=/\v^# / end=/$/ oneline contained contains=dodderMetadateiBezeichnungPrefix,dodderMetadateiBezeichnung

syn match dodderMetadateiComment /\v[^\n]+/ contained contains=@NoSpell
syn match dodderMetadateiCommentPrefix /^%/ contained contains=@NoSpell nextgroup=dodderMetadateiComment
syn region dodderMetadateiCommentRegion start=/^%/ end=/$/ oneline contained contains=dodderMetadateiCommentPrefix,dodderMetadateiComment

syn match dodderMetadateiEtikett /\v[^\n]+/ contained contains=@NoSpell
syn match dodderMetadateiEtikettPrefix /\v^- / contained
syn region dodderMetadateiEtikettRegion start=/\v^- / end=/$/ oneline contained contains=dodderMetadateiEtikett,dodderMetadateiEtikettPrefix

syn match dodderMetadateiAkteBase /\v[^\n]*\.@=/ contained contains=@NoSpell nextgroup=dodderMetadateiAkteDot
syn match dodderMetadateiAkteDot /\v\./ contained contains=@NoSpell nextgroup=dodderMetadateiAkteExt
syn match dodderMetadateiAkteExt /\v\w+/ contained contains=@NoSpell
syn match dodderMetadateiAktePrefix /\v^! / contained nextgroup=dodderMetadateiAkteBase
syn region dodderMetadateiAkteRegion start=/\v^! / end=/$/ oneline contained contains=dodderMetadateiAkte,dodderMetadateiAktePrefix,dodderMetadateiAkteBase,dodderMetadateiAkteExt

" highlight default link dodderLinePrefixRoot Special
highlight default link dodderMetadatei Normal
highlight default link dodderMetadateiBezeichnung Title
highlight default link dodderMetadateiEtikett Constant
highlight default link dodderMetadateiAkteBase Underlined
highlight default link dodderMetadateiAkteExt Type
highlight default link dodderMetadateiComment Comment

" let b:current_syntax = 'dodder-metadatei'
