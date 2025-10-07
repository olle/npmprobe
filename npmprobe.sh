#!/usr/bin/env bash
set -euo pipefail

result=0

file=${1--}
while IFS= read -r package; do

  pkg="${package%@*}"; ver="${package#*@}"
  found=0

  for file in 'package.json' 'package-lock.json'; do

    if out=$(mdfind -0 -name $file 2>/dev/null \
      | xargs -0 grep "\"${pkg}\":" \
      | grep "${ver}\"" 2>/dev/null || true); then

      if [[ -n "$out" ]]; then
        echo "[FOUND] $pkg@$ver in the following $file files:"
        mdfind -0 -name $file 2>/dev/null \
          | xargs -0 grep "\"${pkg}\":" \
          | grep "${ver}\"" \
          | awk '{print "\t"$1}' \
          | sort \
          | uniq
        found=1
        result=1
      fi

    fi
  done

  if [[ $found -eq 0 ]];then
    echo "[OK]   $pkg@$ver not present in any files"
  fi

done < <(cat -- "$file")

exit $result
