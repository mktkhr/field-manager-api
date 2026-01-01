#!/bin/bash

echo "Step Functionsステートマシンを作成しています..."

# IAMロールを作成
awslocal iam create-role \
  --role-name stepfunctions-role \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Service": "states.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
      }
    ]
  }' 2>/dev/null || echo "IAMロールは既に存在します"

# ステートマシン定義を読み込み
cat > /tmp/state-machine.json << 'EOF'
{
  "Comment": "Wagri Import Workflow",
  "StartAt": "FetchFromWagri",
  "States": {
    "FetchFromWagri": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:000000000000:function:wagri-fetcher",
      "Next": "ProcessAndUpsert",
      "Retry": [{
        "ErrorEquals": ["States.TaskFailed"],
        "IntervalSeconds": 30,
        "MaxAttempts": 3,
        "BackoffRate": 2.0
      }],
      "Catch": [{
        "ErrorEquals": ["States.ALL"],
        "Next": "FailState"
      }]
    },
    "ProcessAndUpsert": {
      "Type": "Pass",
      "Comment": "ローカル環境ではEKS Jobの代わりにPassステートを使用。実際の処理はDocker直接実行で行う。",
      "End": true
    },
    "FailState": {
      "Type": "Fail",
      "Error": "WorkflowFailed",
      "Cause": "ワークフローの実行中にエラーが発生しました"
    }
  }
}
EOF

# ステートマシンを作成
awslocal stepfunctions create-state-machine \
  --name "wagri-import-workflow" \
  --definition file:///tmp/state-machine.json \
  --role-arn "arn:aws:iam::000000000000:role/stepfunctions-role" \
  2>/dev/null || echo "ステートマシンは既に存在します"

echo "Step Functionsステートマシン作成完了: wagri-import-workflow"
