import * as cdk from 'aws-cdk-lib';
import {Construct} from 'constructs';

export class TgBotBankStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const domainName = process.env.PUBLIC_API!!;
        const tgToken = process.env.TG_BOT_TOKEN!!;

        // SSL certificate
        const certificate = new cdk.aws_certificatemanager.Certificate(this, 'Certificate', {
            domainName: domainName,
            validation: cdk.aws_certificatemanager.CertificateValidation.fromDns(),
        });

        // DynamoDB
        const balanceTable = new cdk.aws_dynamodb.Table(this, 'Balance', {
            partitionKey: {name: 'id', type: cdk.aws_dynamodb.AttributeType.STRING},
        });

        // SQS
        const queue = new cdk.aws_sqs.Queue(this, 'TgBotQueue', {
            visibilityTimeout: cdk.Duration.seconds(30)
        });

        // First Lambda function
        const listenerApi = new cdk.aws_lambda.Function(this, 'ListenerApi', {
            runtime: cdk.aws_lambda.Runtime.GO_1_X,
            code: cdk.aws_lambda.Code.fromAsset('app/cmd/listener'),
            handler: 'main',
            environment: {
                'PUBLIC_URL': domainName,
                'TG_BOT_TOKEN': tgToken,
                'TABLE_NAME': balanceTable.tableName,
                'QUEUE': queue.queueUrl,
            },
        });
        queue.grantSendMessages(listenerApi);
        balanceTable.grantReadWriteData(listenerApi);

        // Second Lambda function
        const chargerApi = new cdk.aws_lambda.Function(this, 'ChargerApi', {
            runtime: cdk.aws_lambda.Runtime.GO_1_X,
            code: cdk.aws_lambda.Code.fromAsset('app/cmd/charger'),
            handler: 'main',
            environment: {
                'TG_BOT_TOKEN': tgToken,
                'TABLE_NAME': balanceTable.tableName,
            },
            events: [
                new cdk.aws_lambda_event_sources.SqsEventSource(queue)
            ],
        });
        balanceTable.grantReadWriteData(chargerApi);

        // API Gateway
        const gateway = new cdk.aws_apigateway.LambdaRestApi(this, 'Api', {
            handler: listenerApi,
            restApiName: "TgBotApi",
            proxy: true,
            integrationOptions: {
                proxy: true,
                allowTestInvoke: true
            }
        });

        // Domain
        const customDomain = new cdk.aws_apigateway.DomainName(this, 'Domain', {
            domainName: domainName,
            certificate: certificate,
        });
        new cdk.aws_apigateway.BasePathMapping(this, 'BasePathMapping', {
            domainName: customDomain,
            restApi: gateway,
        });
        const hostedZone = cdk.aws_route53.HostedZone.fromLookup(this, 'HostedZone', {
            domainName: domainName,
        });
        new cdk.aws_route53.ARecord(this, 'AliasRecord', {
            zone: hostedZone,
            target: cdk.aws_route53.RecordTarget.fromAlias(new cdk.aws_route53_targets.ApiGatewayDomain(customDomain)),
            recordName: domainName,
        });
        new cdk.CfnOutput(this, 'ApiGatewayUrl', {
            value: gateway.url,
        });
        new cdk.CfnOutput(this, 'CustomDomainUrl', {
            value: `https://${customDomain.domainName}`,
        });
    }
}
