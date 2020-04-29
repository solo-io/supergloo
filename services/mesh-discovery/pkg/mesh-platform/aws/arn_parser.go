package aws

import "github.com/aws/aws-sdk-go/aws/arn"

type arnParser struct{}

func NewArnParser() ArnParser {
	return &arnParser{}
}

func (a *arnParser) ParseAccountID(arnString string) (string, error) {
	parse, err := arn.Parse(arnString)
	if err != nil {
		return "", ARNParseError(err, arnString)
	}
	return parse.AccountID, nil
}
