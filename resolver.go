package main

import (
  "errors"
  "fmt"
)

func (resolver * Resolver) init() error {

  if resolver.ActionOnFail == "hold" {
    resolver.state.onFail_HOLD = true
    return nil
  }

  if resolver.ActionOnFail == "drop" {
    resolver.state.onFail_HOLD = false
    return nil
  }

  msg := fmt.Sprintf("unsupported value \"%s\" for option 'on_fail'", resolver.ActionOnFail)
  return errors.New(msg)

}
