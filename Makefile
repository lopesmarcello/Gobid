APP_NAME=gobid
REGION=us-east-1
VPC_ID=vpc-0723f7078543bf4c1
SG_NAME=gobid-sg
ACCOUNT_ID=292212378373
ECL_URL=$(ACCOUNT_ID).dkr.ecr.$(REGION).amazonaws.com
REPO_URL=$(ECL_URL)/$(APP_NAME)

create-sg:
	@if ! aws ec2 describe-security-groups --filter "Name=group-name, Values=$(SG_NAME)" --region $(REGION) --query "SecurityGroups[*].GroupId" --output text | grep -qE 'sg-'; then \
		echo "Creating security group $(SG_NAME)..."; \
		aws ec2 create-security-group \
			--group-name $(SG_NAME) \
			--description "Allow Postgres for Gobid" \
			--vpc-id $(VPC_ID) \
			--region $(REGION); \
	else \
		echo "Security group $(SG_NAME) already exists."; \
	fi; \
	SG_ID=$$(aws ec2 describe-security-groups --filter "Name=group-name,Values=$(SG_NAME)" --region $(REGION) --query "SecurityGroups[*].GroupId" --output text); \
	echo "Authorizing port 5432 on SG $$SG_ID..."; \
	aws ec2 authorize-security-group-ingress \
		--group-id $$SG_ID \
		--protocol tcp \
		--port 5432 \
		--cidr 0.0.0.0/0 \
		--region $(REGION) || echo "Ingress rule already exists."

create-ecr:
	aws ecr describe-repositories --repository-name $(APP_NAME) --region $(REGION) || \
	aws ecr create-repository --repository-name $(APP_NAME) --region $(REGION)

build:
	docker build -t $(APP_NAME) -f Dockerfile.prod .

tag:
	docker tag ${APP_NAME}:latest


delete-sg:
	@SG_ID=$(shell aws ec2 describe-security-groups --filter "Name=group-name,Values=$(SG_NAME)" --region $(REGION) --query "SecurityGroups[*].GroupId" --output text); \
	if [ -n "$(SG_ID)" ]; then \
		echo "Deleting security group $(SG_NAME) with ID $(SG_ID)..."; \
		aws ec2 delete-security-group --group-id $(SG_ID) --region $(REGION); \
		echo "Security group $(SG_NAME) deleted."; \
	else \
		echo "Security group $(SG_NAME) not found."; \
	fi


