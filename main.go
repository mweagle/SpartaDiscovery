package main

import (
	"encoding/json"
	"fmt"
	gocf "github.com/crewjam/go-cloudformation"
	"net/http"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
)

////////////////////////////////////////////////////////////////////////////////
// Bucket handler
//
func echoS3DynamicBucketEvent(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {

	config, _ := sparta.Discover()
	logger.WithFields(logrus.Fields{
		"RequestID":     context.AWSRequestID,
		"Event":         string(*event),
		"Configuration": config,
	}).Info("Request received")

	fmt.Fprintf(w, string(*event))
}

////////////////////////////////////////////////////////////////////////////////
// Handler registration
//
func appendDynamicS3BucketLambda(lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {

	s3BucketResourceName := sparta.CloudFormationResourceName("S3DynamicBucket")

	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, echoS3DynamicBucketEvent, nil)
	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.S3Permission{
		BasePermission: sparta.BasePermission{
			SourceArn: gocf.Ref(s3BucketResourceName),
		},
		Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
	})
	lambdaFn.DependsOn = append(lambdaFn.DependsOn, s3BucketResourceName)

	// Add permission s.t. the lambda function could read from the S3 bucket
	lambdaFn.RoleDefinition.Privileges = append(lambdaFn.RoleDefinition.Privileges,
		sparta.IAMRolePrivilege{
			Actions:  []string{"s3:GetObject", "s3:HeadObject"},
			Resource: spartaCF.S3AllKeysArnForBucket(gocf.Ref(s3BucketResourceName)),
		})

	lambdaFn.Decorator = func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		template *gocf.Template,
		logger *logrus.Logger) error {
		cfResource := template.AddResource(s3BucketResourceName, &gocf.S3Bucket{
			AccessControl: gocf.String("PublicRead"),
		})
		cfResource.DeletionPolicy = "Delete"
		return nil
	}
	return append(lambdaFunctions, lambdaFn)
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	// Deploy it
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFunctions = appendDynamicS3BucketLambda(lambdaFunctions)

	sparta.Main("SpartaDiscovery",
		fmt.Sprintf("Test sparta.Discover() function"),
		lambdaFunctions,
		nil,
		nil)
}
