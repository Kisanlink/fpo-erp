#!/bin/bash

# Kisanlink ERP - Create New FPO Configuration Script
# Usage: ./create-fpo.sh <fpo-id>

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INFRASTRUCTURE_DIR="$(dirname "$SCRIPT_DIR")"
PARAMETERS_DIR="$INFRASTRUCTURE_DIR/parameters"

# Check arguments
if [ $# -lt 1 ]; then
    echo -e "${RED}Error: Missing required argument${NC}"
    echo "Usage: $0 <fpo-id>"
    echo "Example: $0 fpo002"
    exit 1
fi

FPO_ID=$1

# Validate FPO ID format
if [[ ! $FPO_ID =~ ^fpo[0-9]{3}$ ]]; then
    echo -e "${RED}Error: Invalid FPO ID format${NC}"
    echo "FPO ID must be in format fpo### (e.g., fpo001, fpo002)"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Create New FPO Configuration${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "FPO ID: ${YELLOW}${FPO_ID}${NC}"
echo -e "${GREEN}========================================${NC}"

# Template file
TEMPLATE_FILE="$PARAMETERS_DIR/template.json"

# Output files
STAGING_FILE="$PARAMETERS_DIR/${FPO_ID}-staging.json"
PRODUCTION_FILE="$PARAMETERS_DIR/${FPO_ID}-production.json"

# Check if template exists
if [ ! -f "$TEMPLATE_FILE" ]; then
    echo -e "${RED}Error: Template file not found: $TEMPLATE_FILE${NC}"
    exit 1
fi

# Check if files already exist
if [ -f "$STAGING_FILE" ]; then
    echo -e "${YELLOW}Warning: Staging configuration already exists: $STAGING_FILE${NC}"
    read -p "Overwrite? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo -e "${YELLOW}Skipping staging configuration${NC}"
        STAGING_FILE=""
    fi
fi

if [ -f "$PRODUCTION_FILE" ]; then
    echo -e "${YELLOW}Warning: Production configuration already exists: $PRODUCTION_FILE${NC}"
    read -p "Overwrite? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo -e "${YELLOW}Skipping production configuration${NC}"
        PRODUCTION_FILE=""
    fi
fi

# Create staging configuration
if [ -n "$STAGING_FILE" ]; then
    echo ""
    echo -e "${GREEN}Creating staging configuration...${NC}"

    # Copy template and update FPO ID and environment
    jq --arg fpo "$FPO_ID" \
       '(.[] | select(.ParameterKey == "FPOIdentifier") | .ParameterValue) |= $fpo |
        (.[] | select(.ParameterKey == "Environment") | .ParameterValue) |= "staging" |
        (.[] | select(.ParameterKey == "DesiredCount") | .ParameterValue) |= "1" |
        (.[] | select(.ParameterKey == "TaskCPU") | .ParameterValue) |= "512" |
        (.[] | select(.ParameterKey == "TaskMemory") | .ParameterValue) |= "1024" |
        (.[] | select(.ParameterKey == "DBInstanceClass") | .ParameterValue) |= "db.t3.micro" |
        (.[] | select(.ParameterKey == "DBAllocatedStorage") | .ParameterValue) |= "20" |
        (.[] | select(.ParameterKey == "ECRImageUri") | .ParameterValue) |= "123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:staging-latest" |
        (.[] | select(.ParameterKey == "AAAServiceURL") | .ParameterValue) |= "https://aaa-staging.kisanlink.com" |
        (.[] | select(.ParameterKey == "AAAGrpcAddress") | .ParameterValue) |= "aaa-grpc-staging.kisanlink.com:9090" |
        (.[] | select(.ParameterKey == "CertificateArn") | .ParameterValue) |= ""' \
       "$TEMPLATE_FILE" > "$STAGING_FILE"

    echo -e "${GREEN}✓ Created: $STAGING_FILE${NC}"
fi

# Create production configuration
if [ -n "$PRODUCTION_FILE" ]; then
    echo ""
    echo -e "${GREEN}Creating production configuration...${NC}"

    # Copy template and update FPO ID and environment
    jq --arg fpo "$FPO_ID" \
       '(.[] | select(.ParameterKey == "FPOIdentifier") | .ParameterValue) |= $fpo |
        (.[] | select(.ParameterKey == "Environment") | .ParameterValue) |= "production" |
        (.[] | select(.ParameterKey == "DesiredCount") | .ParameterValue) |= "2" |
        (.[] | select(.ParameterKey == "TaskCPU") | .ParameterValue) |= "1024" |
        (.[] | select(.ParameterKey == "TaskMemory") | .ParameterValue) |= "2048" |
        (.[] | select(.ParameterKey == "DBInstanceClass") | .ParameterValue) |= "db.t3.small" |
        (.[] | select(.ParameterKey == "DBAllocatedStorage") | .ParameterValue) |= "50" |
        (.[] | select(.ParameterKey == "ECRImageUri") | .ParameterValue) |= "123456789012.dkr.ecr.us-west-2.amazonaws.com/kisanlink-erp:production-latest" |
        (.[] | select(.ParameterKey == "AAAServiceURL") | .ParameterValue) |= "https://aaa.kisanlink.com" |
        (.[] | select(.ParameterKey == "AAAGrpcAddress") | .ParameterValue) |= "aaa-grpc.kisanlink.com:9090" |
        (.[] | select(.ParameterKey == "CertificateArn") | .ParameterValue) |= "arn:aws:acm:us-west-2:123456789012:certificate/xxxxx"' \
       "$TEMPLATE_FILE" > "$PRODUCTION_FILE"

    echo -e "${GREEN}✓ Created: $PRODUCTION_FILE${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Configuration Files Created${NC}"
echo -e "${GREEN}========================================${NC}"

if [ -n "$STAGING_FILE" ]; then
    echo -e "Staging:    ${STAGING_FILE}"
fi

if [ -n "$PRODUCTION_FILE" ]; then
    echo -e "Production: ${PRODUCTION_FILE}"
fi

echo ""
echo -e "${YELLOW}⚠ IMPORTANT: Update the following values in the configuration files:${NC}"
echo -e "  - VpcId, PublicSubnet1Id, PublicSubnet2Id, PrivateSubnet1Id, PrivateSubnet2Id"
echo -e "  - ECRImageUri (replace AWS account ID)"
echo -e "  - DBPassword (use a strong 32+ character password)"
echo -e "  - JWTSecret (use a strong 32+ character secret)"
echo -e "  - AAAJWTSecret (use AAA service JWT secret)"
echo -e "  - CertificateArn (for production HTTPS)"
echo ""
echo -e "${GREEN}Next Steps:${NC}"
echo -e "  1. Update configuration values: vi ${PARAMETERS_DIR}/${FPO_ID}-staging.json"
echo -e "  2. Deploy staging: make deploy-fpo FPO=${FPO_ID} ENV=staging"
echo -e "  3. Test staging deployment"
echo -e "  4. Update production config: vi ${PARAMETERS_DIR}/${FPO_ID}-production.json"
echo -e "  5. Deploy production: make deploy-fpo FPO=${FPO_ID} ENV=production"
