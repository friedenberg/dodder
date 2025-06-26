# assert_output_cut
# =============
#
# Summary: Fail if `$output' does not match the expected output.
#
# Usage: assert_output_cut [-p | -e] [- | [--] <expected>]
#
# Options:
#   -p, --partial  Match if `expected` is a substring of `$output`
#   -e, --regexp   Treat `expected` as an extended regular expression
#   -, --stdin     Read `expected` value from STDIN
#   <expected>     The expected value, substring or regular expression
#
# IO:
#   STDIN - [=$1] expected output
#   STDERR - details, on failure
#            error message, on error
# Globals:
#   output
# Returns:
#   0 - if output matches the expected value/partial/regexp
#   1 - otherwise
#
# This function verifies that a command or function produces the expected output.
# (It is the logical complement of `refute_output`.)
# Output matching can be literal (the default), partial or by regular expression.
# The expected output can be specified either by positional argument or read from STDIN by passing the `-`/`--stdin` flag.
#
# ## Literal matching
#
# By default, literal matching is performed.
# The assertion fails if `$output` does not equal the expected output.
#
#   ```bash
#   @test 'assert_output_cut()' {
#     run echo 'have'
#     assert_output_cut 'want'
#   }
#
#   @test 'assert_output_cut() with pipe' {
#     run echo 'hello'
#     echo 'hello' | assert_output_cut -
#   }
#
#   @test 'assert_output_cut() with herestring' {
#     run echo 'hello'
#     assert_output_cut - <<< hello
#   }
#   ```
#
# On failure, the expected and actual output are displayed.
#
#   ```
#   -- output differs --
#   expected : want
#   actual   : have
#   --
#   ```
#
# ## Existence
#
# To assert that any output exists at all, omit the `expected` argument.
#
#   ```bash
#   @test 'assert_output_cut()' {
#     run echo 'have'
#     assert_output_cut
#   }
#   ```
#
# On failure, an error message is displayed.
#
#   ```
#   -- no output --
#   expected non-empty output, but output was empty
#   --
#   ```
#
# ## Partial matching
#
# Partial matching can be enabled with the `--partial` option (`-p` for short).
# When used, the assertion fails if the expected _substring_ is not found in `$output`.
#
#   ```bash
#   @test 'assert_output_cut() partial matching' {
#     run echo 'ERROR: no such file or directory'
#     assert_output_cut --partial 'SUCCESS'
#   }
#   ```
#
# On failure, the substring and the output are displayed.
#
#   ```
#   -- output does not contain substring --
#   substring : SUCCESS
#   output    : ERROR: no such file or directory
#   --
#   ```
#
# ## Regular expression matching
#
# Regular expression matching can be enabled with the `--regexp` option (`-e` for short).
# When used, the assertion fails if the *extended regular expression* does not match `$output`.
#
# *__Note__:
# The anchors `^` and `$` bind to the beginning and the end (respectively) of the entire output;
# not individual lines.*
#
#   ```bash
#   @test 'assert_output_cut() regular expression matching' {
#     run echo 'Foobar 0.1.0'
#     assert_output_cut --regexp '^Foobar v[0-9]+\.[0-9]+\.[0-9]$'
#   }
#   ```
#
# On failure, the regular expression and the output are displayed.
#
#   ```
#   -- regular expression does not match output --
#   regexp : ^Foobar v[0-9]+\.[0-9]+\.[0-9]$
#   output : Foobar 0.1.0
#   --
#   ```
assert_output_cut() {
  local -i is_mode_partial=0
  local -i is_mode_regexp=0
  local -i is_mode_nonempty=0
  local -i use_stdin=0
  local -i should_sort=0
  : "${output?}"

  # Handle options.
  if (( $# == 0 )); then
    is_mode_nonempty=1
  fi

  cut_args=()

  while (( $# > 0 )); do
    case "$1" in
    -p|--partial) is_mode_partial=1; shift ;;
    -e|--regexp) is_mode_regexp=1; shift ;;
    -|--stdin) use_stdin=1; shift ;;
    --) shift; break ;;
    *) cut_args+=("$1"); shift;;
    esac
  done

  while (( $# > 0 )); do
    case "$1" in
    -p|--partial) is_mode_partial=1; shift ;;
    -e|--regexp) is_mode_regexp=1; shift ;;
    -|--stdin) use_stdin=1; shift ;;
    -s|--sort) should_sort=1; shift ;;
    --) shift; break ;;
    *) break ;;
    esac
  done

  if (( is_mode_partial )) && (( is_mode_regexp )); then
    echo "\`--partial' and \`--regexp' are mutually exclusive" \
    | batslib_decorate 'ERROR: assert_output_cut' \
    | fail
    return $?
  fi

  # Arguments.
  local expected
  if (( use_stdin )); then
    expected="$(cat -)"
  else
    expected="${1-}"
  fi

  local expected_sorted output_sorted
  expected_sorted="$(echo -n "$expected" | cut "${cut_args[@]}")"
  output_sorted="$(echo -n "$output" | cut "${cut_args[@]}")"
  # echo "$expected_sorted" >&2
  # echo "$output_sorted" >&2
  # fail
  
  if (( should_sort )); then
    expected_sorted="$(echo -n "$expected_sorted" | sort)"
    output_sorted="$(echo -n "$output_sorted" | sort)"
  fi

  # Matching.
  if (( is_mode_nonempty )); then
    if [ -z "$output_sorted" ]; then
      echo 'expected non-empty output, but output was empty' \
      | batslib_decorate 'no output' \
      | fail
    fi
  elif (( is_mode_regexp )); then
    if [[ '' =~ $expected_sorted ]] || (( $? == 2 )); then
      echo "Invalid extended regular expression: \`$expected_sorted'" \
      | batslib_decorate 'ERROR: assert_output_cut' \
      | fail
    elif ! [[ $output_sorted =~ $expected_sorted ]]; then
      batslib_print_kv_single_or_multi 6 \
      'regexp'  "$expected_sorted" \
      'output' "$output_sorted" \
      | batslib_decorate 'regular expression does not match output' \
      | fail
    fi
  elif (( is_mode_partial )); then
    if [[ $output_sorted != *"$expected_sorted"* ]]; then
      batslib_print_kv_single_or_multi 9 \
      'substring' "$expected_sorted" \
      'output'    "$output_sorted" \
      | batslib_decorate 'output does not contain substring' \
      | fail
    fi
  else
    if [[ $output_sorted != "$expected_sorted" ]]; then
      batslib_print_kv_single_or_multi 8 \
      'expected' "$expected_sorted" \
      'actual'   "$output_sorted" \
      | batslib_decorate 'output differs' \
      | fail
    fi
  fi
}
