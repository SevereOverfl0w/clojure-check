" This Source Code Form is subject to the terms of the Mozilla Public
" License, v. 2.0. If a copy of the MPL was not distributed with this
" file, You can obtain one at http://mozilla.org/MPL/2.0/.

if exists('g:loaded_clojure_check')
  finish
endif
let g:loaded_clojure_check = 1

let s:check_version = '0.2'
let s:base_dir = expand('<sfile>:h:h')
let g:clojure_check_bin = s:base_dir.'/bin/clojure-check-v'.s:check_version

function! s:ClojureHost()
  return fireplace#client().connection.transport.host
endfunction

function! s:ClojurePort()
  return fireplace#client().connection.transport.port
endfunction

function! s:ClojureHostPort()
  if exists("g:acid_loaded")
    let host_port = AcidGetUrl()
    if host_port == v:null
      throw 'Acid: No repl connection'
    endif
  else
    let host_port = [s:ClojureHost(), s:ClojurePort()]
  endif
  return join(host_port, ":")
endfunction

function! s:ClojureNs()
  if exists("g:acid_loaded")
    return AcidGetNs()
  else
    return fireplace#ns(a:buffer)
  endif
endfunction

function! ClojureCheckArgs(buffer)
  return ['-nrepl', s:ClojureHostPort(), '-namespace', s:ClojureNs()]
endfunction

function! ClojureCheck(buffer)
  try
    return g:clojure_check_bin.' '.join(ClojureCheckArgs(a:buffer) + ['-file', '-'], ' ')
  catch /Fireplace|Acid/
    return ''
  endtry
endfunction

try
  call ale#linter#Define('clojure', {
  \   'name': 'clojure_check',
  \   'executable': g:clojure_check_bin,
  \   'command_callback': 'ClojureCheck',
  \   'callback': 'ale#handlers#HandleUnixFormatAsError',
  \})
catch /E117/
endtry

let g:neomake_clojure_check_maker = {
    \ 'exe': g:clojure_check_bin,
    \ 'errorformat': '%f:%l:%c: %m',
    \ }

function! g:neomake_clojure_check_maker.args()
  try
    return ClojureCheckArgs(bufnr('%'))+ ['-file']
  catch /Acid/
    return []
  endtry
endfunction
