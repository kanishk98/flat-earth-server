package main

import (
	"math/big"

	"github.com/zclconf/go-cty/cty"
)

func getCtyValue(ctyType string, value interface{}) cty.Value {
	switch ctyType {
	case "string":
		return cty.StringVal(value.(string))
	case "bool":
		return cty.BoolVal(value.(bool))
	case "number":
		return cty.NumberVal(value.(*big.Float))
	}
	return cty.ObjectVal(value.(map[string]cty.Value))
}
