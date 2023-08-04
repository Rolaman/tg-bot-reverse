# Telegram Bot AWS Task

You need to write up a Telegram Bot in Go. The bot reverses user messages and charges the user for each message. A serverless approach has to be used, 2 lambdas, first lambda for incoming messages that simply checks that the message is a text message and that the request is from telegram, it may also be the one creating the user. The second lambda is the one that actually charges the user for the message. The cost of the message is decided by a function that randomly generates a number between 0-100 in the second lambda, the cost is decided dynamically for each message.
Be carefull, as you cannot simply charge the user in the second Lambda as this could lead to reentrancy problem, double spend. But you also can't charge the user exactly on the first lambda since the cost is not known there, it is only known in the second lambda.
Think about your approach to handle these cases.

To topup the bot we should be able to call: /topup 100
To check the balance we should be able to call: /balance
Bot should only accept text messages

You need to use the AWS CDK + Go as the stack, use DynamoDB as the user (and user balances) store.


# Welcome to your CDK TypeScript project

This is a blank project for CDK development with TypeScript.

The `cdk.json` file tells the CDK Toolkit how to execute your app.

## Useful commands

* `npm run build`   compile typescript to js
* `npm run watch`   watch for changes and compile
* `npm run test`    perform the jest unit tests
* `cdk deploy`      deploy this stack to your default AWS account/region
* `cdk diff`        compare deployed stack with current state
* `cdk synth`       emits the synthesized CloudFormation template
