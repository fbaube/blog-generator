package generator

import (
  "fmt"
  "html/template"
  SU "github.com/fbaube/stringutils"
)

type BaseConfig struct {
     *template.Template
     Dest string
     BlogProps SU.PropSet // *IndexWriter
}

func (bc *BaseConfig) String() string {
  return fmt.Sprintf("BasCgf:HasTmpl?<%t>Dest<%s>Props<%v>",
    (nil != bc.Template), bc.Dest, bc.BlogProps)
}
