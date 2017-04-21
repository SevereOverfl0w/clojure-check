" This Source Code Form is subject to the terms of the Mozilla Public
" License, v. 2.0. If a copy of the MPL was not distributed with this
" file, You can obtain one at http://mozilla.org/MPL/2.0/.

if exists('g:loaded_clojure_check')
  finish
endif
let g:loaded_clojure_check = 1

let s:check_version = '0.1'
let s:base_dir = expand('<sfile>:h:h')
" let s:clojure_check_bin = s:base_dir.'/bin/clojure-check-v'.s:check_version
let s:clojure_check_bin = s:base_dir.'clojure-check'

function! s:ClojureHost()
  return fireplace#client().connection.transport.host
endfunction

function! s:ClojurePort()
  return fireplace#client().connection.transport.port
endfunction

function! ClojureCheck(buffer)
  try
    return s:clojure_check_bin.' '.s:ClojureHost().':'.s:ClojurePort().' '.fireplace#ns(a:buffer)
  catch /Fireplace/
    return ''
  endtry
endfunction

try
  call ale#linter#Define('clojure', {
  \   'name': 'clojure_check',
  \   'executable': s:clojure_check_bin,
  \   'command_callback': 'ClojureCheck',
  \   'callback': 'ale#handlers#HandleUnixFormatAsError',
  \})
catch /E117/
endtry
