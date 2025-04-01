function assert() {
    output=$(eval "$1")
    if [[ $output != "$2" ]]; then
      echo "Test failed: expected '$2', but got '$output'"
      exit 1
    fi
}

function assert_eventually() {
    max_attempts=20
    for ((i=1; i<=max_attempts; i++)); do
      output=$(eval "$1")
      if [[ $output == "$2" ]]; then
        echo "Attempt $i succeeded"
        return 0
      else
        echo "Attempt $i failed - expected '$2', but got '$output'"
      fi
      sleep 5
    done
    echo "Assert failed after $max_attempts attempts"
    exit 1
}
