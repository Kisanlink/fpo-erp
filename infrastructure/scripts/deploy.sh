#!/bin/bash

# Kisanlink ERP - CloudFormation Deployment Script
# Usage: ./deploy.sh <fpo-id> <environment> [--no-confirm]

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
PARAMETERS_DIR="$INFRASTRUCTURE_DIR/parameters"

# Check arguments
if [ $# -lt 2 ]; then
    echo -e "${RED}Error: Missing required arguments${NC}"
    echo "Usage: $0 <fpo-id> <environment> [--no-confirm]"
    echo "Example: $0 fpo001 staging"
    echo "Example: $0 fpo002 production --no-confirm"
    exit 1
fi

FPO_ID=$1
ENVIRONMENT=$2
NO_CONFIRM=$3

# Validate FPO ID format
if [[ ! $FPO_ID =~ ^fpo[0-9]{3}$ ]]; then
    echo -e "${RED}Error: Invalid FPO ID format${NC}"
    echo "FPO ID must be in format fpo### (e.g., fpo001, fpo002)"
    exit 1
fi

# Validate environment
if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
    echo -e "${RED}Error: Invalid environment${NC}"
    echo "Environment must be 'staging' or 'production'"
    exit 1
fi

# Stack name
STACK_NAME="kisanlink-erp-${FPO_ID}-${ENVIRONMENT}"

# Template and parameter files
TEMPLATE_FILE="$TEMPLATES_DIR/erp-application.yaml"
PARAMETER_FILE="$PARAMETERS_DIR/${FPO_ID}-${ENVIRONMENT}.json"

# Check if files exist
if [ ! -f "$TEMPLATE_FILE" ]; then
    echo -e "${RED}Error: Template file not found: $TEMPLATE_FILE${NC}"
    exit 1
fi

if [ ! -f "$PARAMETER_FILE" ]; then
    echo -e "${RED}Error: Parameter file not found: $PARAMETER_FILE${NC}"
    echo -e "${YELLOW}Tip: Copy template.json to ${FPO_ID}-${ENVIRONMENT}.json and update values${NC}"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Kisanlink ERP Deployment${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "FPO ID:      ${YELLOW}${FPO_ID}${NC}"
echo -e "Environment: ${YELLOW}${ENVIRONMENT}${NC}"
echo -e "Stack Name:  ${YELLOW}${STACK_NAME}${NC}"
echo -e "Template:    ${TEMPLATE_FILE}"
echo -e "Parameters:  ${PARAMETER_FILE}"
echo -e "${GREEN}========================================${NC}"

# Confirm deployment unless --no-confirm is passed
if [ "$NO_CONFIRM" != "--no-confirm" ]; then
    echo ""
    read -p "Do you want to proceed with deployment? (yes/no): " -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo -e "${YELLOW}Deployment cancelled${NC}"
        exit 0
    fi
fi

echo -e "${GREEN}Step 1: Validating CloudFormation template...${NC}"
if ! aws cloudformation validate-template \
    --template-body file://"$TEMPLATE_FILE" > /dev/null 2>&1; then
    echo -e "${RED}Error: Template validation failed${NC}"
    aws cloudformation validate-template --template-body file://"$TEMPLATE_FILE"
    exit 1
fi
echo -e "${GREEN}✓ Template validation passed${NC}"

echo ""
echo -e "${GREEN}Step 2: Checking if stack exists...${NC}"
STACK_EXISTS=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" 2>&1 || echo "not-found")

if echo "$STACK_EXISTS" | grep -q "does not exist"; then
    echo -e "${YELLOW}Stack does not exist, creating new stack...${NC}"
    ACTION="create"
else
    echo -e "${YELLOW}Stack exists, updating stack...${NC}"
    ACTION="update"
fi

echo ""
if [ "$ACTION" == "create" ]; then
    echo -e "${GREEN}Step 3: Creating CloudFormation stack...${NC}"
    aws cloudformation create-stack \
        --stack-name "$STACK_NAME" \
        --template-body file://"$TEMPLATE_FILE" \
        --parameters file://"$PARAMETER_FILE" \
        --capabilities CAPABILITY_NAMED_IAM \
        --tags \
            Key=FPO,Value="$FPO_ID" \
            Key=Environment,Value="$ENVIRONMENT" \
            Key=ManagedBy,Value=CloudFormation \
            Key=Application,Value=KisanlinkERP

    echo ""
    echo -e "${GREEN}Stack creation initiated. Waiting for completion...${NC}"
    aws cloudformation wait stack-create-complete --stack-name "$STACK_NAME"

else
    echo -e "${GREEN}Step 3: Updating CloudFormation stack...${NC}"
    UPDATE_OUTPUT=$(aws cloudformation update-stack \
        --stack-name "$STACK_NAME" \
        --template-body file://"$TEMPLATE_FILE" \
        --parameters file://"$PARAMETER_FILE" \
        --capabilities CAPABILITY_NAMED_IAM \
        --tags \
            Key=FPO,Value="$FPO_ID" \
            Key=Environment,Value="$ENVIRONMENT" \
            Key=ManagedBy,Value=CloudFormation \
            Key=Application,Value=KisanlinkERP 2>&1 || echo "no-updates")

    if echo "$UPDATE_OUTPUT" | grep -q "No updates are to be performed"; then
        echo -e "${YELLOW}No updates to perform. Stack is already up to date.${NC}"
    else
        echo ""
        echo -e "${GREEN}Stack update initiated. Waiting for completion...${NC}"
        aws cloudformation wait stack-update-complete --stack-name "$STACK_NAME"
    fi
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment Status${NC}"
echo -e "${GREEN}========================================${NC}"

# Get stack outputs
OUTPUTS=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query 'Stacks[0].Outputs' --output table)

echo "$OUTPUTS"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ Deployment Complete!${NC}"
echo -e "${GREEN}========================================${NC}"

# Extract and display Load Balancer URL
LB_URL=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" \
    --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
    --output text)

if [ -n "$LB_URL" ]; then
    echo -e "Application URL: ${YELLOW}${LB_URL}${NC}"
    echo -e "Health Check:    ${YELLOW}${LB_URL}/health${NC}"
fi

echo ""
echo -e "${YELLOW}Note: It may take a few minutes for the ECS service to start and pass health checks.${NC}"
