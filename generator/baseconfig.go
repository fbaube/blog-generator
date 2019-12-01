package generator

import (
  "fmt"
  "html/template"
)

type BaseConfig struct {
     /*Template*/   *template.Template
     Dest/*ination*/ string
     /*Writer*/     *IndexWriter
}

func (bc *BaseConfig) String() string {
  return fmt.Sprintf("BasCgf:(Tmpl==nil)<%t>Dest<%s>IdxWrtr<%s>",
    (nil == bc.Template), bc.Dest, *bc.IndexWriter)
}
