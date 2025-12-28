#!/bin/bash

echo "S3バケットを作成しています..."
awslocal s3 mb s3://field-manager-imports
echo "S3バケット作成完了: field-manager-imports"
