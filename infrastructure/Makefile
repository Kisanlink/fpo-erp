.PHONY: help validate deploy deploy-fpo create-fpo list-stacks describe-stack delete-stack logs shell

# Default target
help:
	@echo "========================================="
	@echo "Kisanlink ERP Infrastructure - Makefile"
	@echo "========================================="
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "  validate           - Validate CloudFormation template"
	@echo "  create-fpo FPO=xxx - Create parameter files for new FPO"
	@echo "  deploy-fpo FPO=xxx ENV=yyy - Deploy ERP for specific FPO and environment"
	@echo "  list-stacks        - List all ERP CloudFormation stacks"
	@echo "  describe-stack STACK=xxx - Describe specific stack"
	@echo "  delete-stack FPO=xxx ENV=yyy - Delete specific stack"
	@echo "  logs FPO=xxx ENV=yyy - Tail CloudWatch logs"
	@echo ""
	@echo "Examples:"
	@echo "  make validate"
	@echo "  make create-fpo FPO=fpo001"
	@echo "  make deploy-fpo FPO=fpo001 ENV=staging"
	@echo "  make deploy-fpo FPO=fpo002 ENV=production"
	@echo "  make list-stacks"
	@echo "  make describe-stack STACK=kisanlink-erp-fpo001-staging"
	@echo "  make delete-stack FPO=fpo001 ENV=staging"
	@echo "  make logs FPO=fpo001 ENV=staging"
	@echo ""

# Validate CloudFormation template
validate:
	@echo "Validating CloudFormation template..."
	@bash scripts/validate.sh

# Create parameter files for new FPO
create-fpo:
ifndef FPO
	@echo "Error: FPO parameter is required"
	@echo "Usage: make create-fpo FPO=fpo001"
	@exit 1
endif
	@echo "Creating configuration for FPO: $(FPO)"
	@bash scripts/create-fpo.sh $(FPO)

# Deploy ERP for specific FPO and environment
deploy-fpo:
ifndef FPO
	@echo "Error: FPO parameter is required"
	@echo "Usage: make deploy-fpo FPO=fpo001 ENV=staging"
	@exit 1
endif
ifndef ENV
	@echo "Error: ENV parameter is required"
	@echo "Usage: make deploy-fpo FPO=fpo001 ENV=staging"
	@exit 1
endif
	@echo "Deploying ERP for FPO: $(FPO), Environment: $(ENV)"
	@bash scripts/deploy.sh $(FPO) $(ENV)

# Deploy with no confirmation (for CI/CD)
deploy-fpo-ci:
ifndef FPO
	@echo "Error: FPO parameter is required"
	@exit 1
endif
ifndef ENV
	@echo "Error: ENV parameter is required"
	@exit 1
endif
	@echo "Deploying ERP for FPO: $(FPO), Environment: $(ENV) (CI/CD mode)"
	@bash scripts/deploy.sh $(FPO) $(ENV) --no-confirm

# List all ERP CloudFormation stacks
list-stacks:
	@echo "Listing all Kisanlink ERP CloudFormation stacks..."
	@aws cloudformation list-stacks \
		--stack-status-filter CREATE_COMPLETE UPDATE_COMPLETE \
		--query 'StackSummaries[?starts_with(StackName, `kisanlink-erp-`)].{Name:StackName,Status:StackStatus,Created:CreationTime}' \
		--output table

# Describe specific stack
describe-stack:
ifndef STACK
	@echo "Error: STACK parameter is required"
	@echo "Usage: make describe-stack STACK=kisanlink-erp-fpo001-staging"
	@exit 1
endif
	@echo "Describing stack: $(STACK)"
	@aws cloudformation describe-stacks --stack-name $(STACK)

# Delete specific stack
delete-stack:
ifndef FPO
	@echo "Error: FPO parameter is required"
	@echo "Usage: make delete-stack FPO=fpo001 ENV=staging"
	@exit 1
endif
ifndef ENV
	@echo "Error: ENV parameter is required"
	@echo "Usage: make delete-stack FPO=fpo001 ENV=staging"
	@exit 1
endif
	@echo "WARNING: This will delete the stack kisanlink-erp-$(FPO)-$(ENV)"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		aws cloudformation delete-stack --stack-name kisanlink-erp-$(FPO)-$(ENV); \
		echo "Stack deletion initiated. Waiting for completion..."; \
		aws cloudformation wait stack-delete-complete --stack-name kisanlink-erp-$(FPO)-$(ENV); \
		echo "Stack deleted successfully."; \
	else \
		echo "Deletion cancelled."; \
	fi

# Tail CloudWatch logs
logs:
ifndef FPO
	@echo "Error: FPO parameter is required"
	@echo "Usage: make logs FPO=fpo001 ENV=staging"
	@exit 1
endif
ifndef ENV
	@echo "Error: ENV parameter is required"
	@echo "Usage: make logs FPO=fpo001 ENV=staging"
	@exit 1
endif
	@echo "Tailing logs for $(FPO) $(ENV)..."
	@aws logs tail /ecs/$(FPO)-$(ENV)-erp --follow

# Get stack outputs
outputs:
ifndef FPO
	@echo "Error: FPO parameter is required"
	@echo "Usage: make outputs FPO=fpo001 ENV=staging"
	@exit 1
endif
ifndef ENV
	@echo "Error: ENV parameter is required"
	@echo "Usage: make outputs FPO=fpo001 ENV=staging"
	@exit 1
endif
	@echo "Stack outputs for kisanlink-erp-$(FPO)-$(ENV):"
	@aws cloudformation describe-stacks --stack-name kisanlink-erp-$(FPO)-$(ENV) \
		--query 'Stacks[0].Outputs' --output table

# Get stack events (for troubleshooting)
events:
ifndef FPO
	@echo "Error: FPO parameter is required"
	@echo "Usage: make events FPO=fpo001 ENV=staging"
	@exit 1
endif
ifndef ENV
	@echo "Error: ENV parameter is required"
	@echo "Usage: make events FPO=fpo001 ENV=staging"
	@exit 1
endif
	@echo "Recent events for kisanlink-erp-$(FPO)-$(ENV):"
	@aws cloudformation describe-stack-events --stack-name kisanlink-erp-$(FPO)-$(ENV) \
		--max-items 20 --query 'StackEvents[*].{Time:Timestamp,Status:ResourceStatus,Type:ResourceType,Reason:ResourceStatusReason}' \
		--output table
