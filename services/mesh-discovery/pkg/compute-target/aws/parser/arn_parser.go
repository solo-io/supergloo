package aws_utils

import (
	"github.com/aws/aws-sdk-go/aws/arn"
)

type arnParser struct{}

func NewArnParser() ArnParser {
	return &arnParser{}
}

func (a *arnParser) parse(arnString string) (*arn.ARN, error) {
	parsedARN, err := arn.Parse(arnString)
	if err != nil {
		return nil, ARNParseError(err, arnString)
	}
	return &parsedARN, nil
}

func (a *arnParser) ParseAccountID(arnString string) (string, error) {
	parsedARN, err := a.parse(arnString)
	if err != nil {
		return "", err
	}
	return parsedARN.AccountID, nil
}

func (a *arnParser) ParseRegion(arnString string) (string, error) {
	parsedARN, err := a.parse(arnString)
	if err != nil {
		return "", err
	}
	return parsedARN.Region, nil
}
