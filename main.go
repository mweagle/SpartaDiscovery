package main

import (
	"context"
	"fmt"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	awsLambdaContext "github.com/aws/aws-lambda-go/lambdacontext"
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////
// Bucket handler
//
func echoS3DynamicBucketEvent(ctx context.Context,
	s3Event awsLambdaEvents.S3Event) (awsLambdaEvents.S3Event, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
	if loggerOk {
		logger.Info("Access structured logger")
	}
	awsContext, _ := awsLambdaContext.FromContext(ctx)

	config, _ := sparta.Discover()
	logger.WithFields(logrus.Fields{
		"RequestID":     awsContext.AwsRequestID,
		"Event":         s3Event,
		"Configuration": config,
	}).Info("Request received")
	return s3Event, nil
}

////////////////////////////////////////////////////////////////////////////////
// Handler registration
//
func appendDynamicS3BucketLambda(lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {

	s3BucketResourceName := sparta.CloudFormationResourceName("S3DynamicBucket")

	lambdaFn := sparta.HandleAWSLambda("echo S3 event",
		echoS3DynamicBucketEvent,
		sparta.IAMRoleDefinition{})

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
		},
	)

	lambdaFn.Decorator = func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		template *gocf.Template,
		context map[string]interface{},
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
