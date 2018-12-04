package appmesh_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/supergloo/pkg/translator/appmesh"
)

var _ = Describe("Appmesh", func() {
	It("works", func() {

		sess, err := session.NewSession(aws.NewConfig().
			WithCredentials(credentials.NewSharedCredentials("", "")))
		Expect(err).NotTo(HaveOccurred())
		svc := appmesh.New(sess, &aws.Config{Region: aws.String("us-east-1")})

		s := NewSyncer()
		err = s.Try(svc)
		Expect(err).NotTo(HaveOccurred())

	})
})
