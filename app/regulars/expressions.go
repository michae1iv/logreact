package regulars

import (
	"regexp"
)

//*This package provides compiled regular expression which are used for checking rules

var (
	//Regulars for checking condition
	Token_reg   *regexp.Regexp
	Value_reg   *regexp.Regexp
	Message_reg *regexp.Regexp
)

func CompileAllExpressions() error {
	var err error

	Token_reg, err = regexp.Compile(`([A-Za-z0-9_.$]+|'(.*?)'|\[(.*?)\]|==|!=|=|:|->|!->|\(|\))`)
	if err != nil {
		return err
	}

	Value_reg, err = regexp.Compile(`'(.*?)'`)
	if err != nil {
		return err
	}

	Message_reg, err = regexp.Compile(`%(.*?)%`)
	if err != nil {
		return err
	}

	return nil
}
