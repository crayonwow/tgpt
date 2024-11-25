up-test-env: down-test-env 
	docker-compose -f ./docker-compose_test.yml build 
	docker-compose -f ./docker-compose_test.yml up -d 
	
down-test-env:
	docker-compose -f ./docker-compose_test.yml down -v --remove-orphans
