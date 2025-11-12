#!/bin/bash

# Kisanlink ERP - CloudFormation Template Validation Script
# Usage: ./validate.sh

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

TEMPLATE_FILE="$TEMPLATES_DIR/erp-application.yaml"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}CloudFormation Template Validation${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Template: ${TEMPLATE_FILE}"
echo -e "${GREEN}========================================${NC}"

if [ ! -f "$TEMPLATE_FILE" ]; then
    echo -e "${RED}Error: Template file not found: $TEMPLATE_FILE${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}Validating template syntax and structure...${NC}"

# Validate template
VALIDATION_OUTPUT=$(aws cloudformation validate-template \
    --template-body file://"$TEMPLATE_FILE" 2>&1)

VALIDATION_EXIT_CODE=$?

if [ $VALIDATION_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ Template validation passed${NC}"
    echo ""
    echo -e "${GREEN}Template Details:${NC}"
    echo "$VALIDATION_OUTPUT" | jq '.'
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Template is valid and ready to deploy${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}✗ Template validation failed${NC}"
    echo ""
    echo -e "${RED}Error Details:${NC}"
    echo "$VALIDATION_OUTPUT"
    echo ""
    exit 1
fi
