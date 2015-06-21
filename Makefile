test:
	go test && node cf.js

launch: test httpd-sim.json
	aws cloudformation create-stack \
	  --stack-name `cat httpd-sim.json | jq -r '.Resources.Settings.Properties.Tags[1].Value'` \
	  --template-body file://httpd-sim.json \
	  --capabilities CAPABILITY_IAM