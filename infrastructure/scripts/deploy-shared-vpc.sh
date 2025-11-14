#!/bin/bash

# Kisanlink ERP - Deploy Shared VPC
# Usage: ./deploy-shared-vpc.sh <environment> [vpc-cidr] [public-subnet-1-cidr] [public-subnet-2-cidr] [private-subnet-1-cidr] [private-subnet-2-cidr] [retention-days]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INFRASTRUCTURE_DIR="$(dirname "$SCRIPT_DIR")"
TEMPLATES_DIR="$INFRASTRUCTURE_DIR/templates"

# Check arguments
if [ $# -lt 1 ]; then
    echo -e "${RED}Error: Missing required arguments${NC}"
    echo "Usage: $0 <environment> [vpc-cidr] [public-subnet-1-cidr] [public-subnet-2-cidr] [private-subnet-1-cidr] [private-subnet-2-cidr] [retention-days]"
    echo "Example: $0 staging"
    echo "Example: $0 production 10.0.0.0/16 10.0.1.0/24 10.0.2.0/24 10.0.11.0/24 10.0.12.0/24 90"
    exit 1
fi

ENVIRONMENT=$1
VPC_CIDR=${2:-"10.0.0.0/16"}
PUBLIC_SUBNET_1_CIDR=${3:-"10.0.1.0/24"}
PUBLIC_SUBNET_2_CIDR=${4:-"10.0.2.0/24"}
PRIVATE_SUBNET_1_CIDR=${5:-"10.0.11.0/24"}
PRIVATE_SUBNET_2_CIDR=${6:-"10.0.12.0/24"}
RETENTION_DAYS=${7:-"90"}

# Validate environment
if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
    echo -e "${RED}Error: Invalid environment${NC}"
    echo "Environment must be 'staging' or 'production'"
    exit 1
fi

# Stack name
STACK_NAME="kisanlink-erp-shared-vpc-${ENVIRONMENT}"

# Template file
TEMPLATE_FILE="$TEMPLATES_DIR/shared-vpc.yaml"

# Check if template exists
if [ ! -f "$TEMPLATE_FILE" ]; then
    echo -e "${RED}Error: Template file not found: $TEMPLATE_FILE${NC}"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Kisanlink ERP - Deploy Shared VPC${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Environment:        $ENVIRONMENT"
echo "VPC CIDR:           $VPC_CIDR"
echo "Public Subnet 1:    $PUBLIC_SUBNET_1_CIDR"
echo "Public Subnet 2:    $PUBLIC_SUBNET_2_CIDR"
echo "Private Subnet 1:   $PRIVATE_SUBNET_1_CIDR"
echo "Private Subnet 2:   $PRIVATE_SUBNET_2_CIDR"
echo "Retention Days:     $RETENTION_DAYS"
echo "Stack Name:         $STACK_NAME"
echo ""

# Check if stack exists
if aws cloudformation describe-stacks --stack-name "$STACK_NAME" &>/dev/null; then
    echo -e "${YELLOW}Stack already exists. Updating...${NC}"
    OPERATION="update"
else
    echo -e "${GREEN}Stack does not exist. Creating...${NC}"
    OPERATION="create"
fi

# Validate template
echo -e "${GREEN}Validating CloudFormation template...${NC}"
aws cloudformation validate-template \
    --template-body file://"$TEMPLATE_FILE" \
    > /dev/null

if [ $? -ne 0 ]; then
    echo -e "${RED}Template validation failed${NC}"
    exit 1
fi

echo -e "${GREEN}Template is valid${NC}"
echo ""

# Deploy stack
echo -e "${GREEN}Deploying CloudFormation stack...${NC}"
aws cloudformation $OPERATION-stack \
    --stack-name "$STACK_NAME" \
    --template-body file://"$TEMPLATE_FILE" \
    --parameters \
        ParameterKey=Environment,ParameterValue="$ENVIRONMENT" \
        ParameterKey=VpcCIDR,ParameterValue="$VPC_CIDR" \
        ParameterKey=PublicSubnet1CIDR,ParameterValue="$PUBLIC_SUBNET_1_CIDR" \
        ParameterKey=PublicSubnet2CIDR,ParameterValue="$PUBLIC_SUBNET_2_CIDR" \
        ParameterKey=PrivateSubnet1CIDR,ParameterValue="$PRIVATE_SUBNET_1_CIDR" \
        ParameterKey=PrivateSubnet2CIDR,ParameterValue="$PRIVATE_SUBNET_2_CIDR" \
        ParameterKey=RetentionInDays,ParameterValue="$RETENTION_DAYS" \
    --tags \
        Key=Name,Value="kisanlink-erp-shared-vpc-${ENVIRONMENT}" \
        Key=Environment,Value="$ENVIRONMENT" \
        Key=Purpose,Value=SharedVPC \
        Key=ManagedBy,Value=CloudFormation

if [ $? -ne 0 ]; then
    echo -e "${RED}Stack ${OPERATION} failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}Waiting for stack ${OPERATION} to complete...${NC}"
aws cloudformation wait stack-${OPERATION}-complete --stack-name "$STACK_NAME"

if [ $? -ne 0 ]; then
    echo -e "${RED}Stack ${OPERATION} failed${NC}"
    echo "Check CloudFormation console for details"
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Stack ${OPERATION} completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Get outputs
echo -e "${GREEN}Stack Outputs (save these for FPO deployments):${NC}"
aws cloudformation describe-stacks \
    --stack-name "$STACK_NAME" \
    --query 'Stacks[0].Outputs[*].[OutputKey,OutputValue]' \
    --output table

echo ""
echo -e "${GREEN}Next steps:${NC}"
echo "1. Deploy FPO stacks using the VPC and Subnet IDs from outputs above"
echo "   (VPC Flow Logs are already configured in this stack)"
echo ""

