#!/usr/bin/env node

var fs = require('fs');
var traverse = require('traverse');

var obj = JSON.parse(fs.readFileSync('httpd.json', 'utf8'));

var USER_PARAMS = {
  "Cluster":            "convox-charlie",
  "Kernel":             "arn:aws:lambda:us-east-1:901416387788:function:convox-formation-convox-charlie-vpc-29bcbf4c",
  "Subnets":            ["subnet-dcc597ab", "subnet-764fd65d", "subnet-7e403727"],
  "VPC":                "vpc-29bcbf4c",

  "Check":              "HTTP:53081/",
  "WebImage":           "docker.io/httpd",
  "WebPort80Balancer":  "80",
  "WebPort80Host":      "53081"
}

var PSEUDO_PARAMS = {
  "AWS::AccountId":        "123456789012",
  "AWS::NotificationARNs": ["arn1", "arn2"],
  "AWS::NoValue":          "",
  "AWS::Region":           "us-west-2",
  "AWS::StackId":          "arn:aws:cloudformation:us-west-2:123456789012:stack/teststack/51af3dc0-da77-11e4-872e-1234567db123",
  "AWS::StackName":        "httpd-" + parseInt(new Date() / 1000),
}

// Replace "Default" values with USER_PARAM if set
// Save all PARAMS for reference eval

var PARAMS = {}

for (var key in obj["Parameters"]) {
  if (USER_PARAMS.hasOwnProperty(key))
    obj["Parameters"][key]["Default"] = USER_PARAMS[key]

  PARAMS[key] = obj["Parameters"][key]["Default"]

  if (Array.isArray(obj["Parameters"][key]["Default"]))
    obj["Parameters"][key]["Default"] = PARAMS[key].join(",")
}

// Evaluate all Parameter Refs
traverse(obj).forEach(function (x) {
  if (typeof(x) == 'object' && Object.keys(x).length == 1 && Object.keys(x)[0] == "Ref") {
    if (PSEUDO_PARAMS.hasOwnProperty(x["Ref"]))
      this.update(PSEUDO_PARAMS[x["Ref"]])

    if (PARAMS.hasOwnProperty(x["Ref"]))
      this.update(PARAMS[x["Ref"]])
  }
})

// Evaluate all Resource Refs
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-ref.html


console.log(JSON.stringify(obj, null, 2))
fs.writeFileSync('httpd-sim.json', JSON.stringify(obj, null, 2))
