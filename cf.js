#!/usr/bin/env node

var assert = require('assert')
var fs = require('fs')
var traverse = require('traverse')

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
  // "AWS::NoValue":          "",
  "AWS::Region":           "us-east-1",
  "AWS::StackId":          "arn:aws:cloudformation:us-east-1:123456789012:stack/teststack/51af3dc0-da77-11e4-872e-1234567db123",
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

// Evaluate all Conditions
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/conditions-section-structure.html

var CONDITIONS = {}

for (var key in obj["Conditions"]) {
  c = obj["Conditions"][key]

  if (c.hasOwnProperty("Fn::Equals")) {
    CONDITIONS[key] = c["Fn::Equals"][0] == c["Fn::Equals"][1]
  }
}

// Evaluate Fn::If

traverse(obj).forEach(function (x) {
  if (typeof(x) == 'object' && Object.keys(x).length == 1 && Object.keys(x)[0] == "Fn::If") {
    cond = x["Fn::If"][0]

    v = CONDITIONS[cond] ? x["Fn::If"][1] : x["Fn::If"][2]

    this.update(v)
  }
})

// Evaluate Fn::Join

traverse(obj).forEach(function (x) {
  if (typeof(x) == 'object' && Object.keys(x).length == 1 && Object.keys(x)[0] == "Fn::Join") {
    c = x["Fn::Join"][0]
    a = x["Fn::Join"][1]

    if (a.every(function(e, i, a) { return typeof e == 'string' }))
      this.update(a.join(c))
  }
})

// Write Simulated JSON

// console.log(JSON.stringify(obj, null, 2))
fs.writeFileSync('httpd-sim.json', JSON.stringify(obj, null, 2))

// Verify Simulated JSON

s = obj["Resources"]["Service"]["Properties"]

assert.equal(s["Cluster"], "convox-charlie")
assert.equal(s["DesiredCount"], "1")
assert.deepEqual(s["LoadBalancers"], [ { "ContainerName": "web", "ContainerPort": "80", "LoadBalancerName": { "Ref": "Balancer" } } ])
assert.equal(s["Name"], null)
assert.deepEqual(s["Role"], {"Ref": "ServiceRole"}) // TODO: Verify ServiceRole Properties
assert.deepEqual(s["TaskDefinition"], {"Ref": "TaskDefinition"})

td = obj["Resources"]["TaskDefinition"]["Properties"]

v = td["Volumes"]

assert.deepEqual(v[0], { "Name": "Docker", "Host": { "SourcePath": "/var/run/docker.sock" } })

t0 = td["ContainerDefinitions"][0]
t1 = td["ContainerDefinitions"][1]

assert.deepEqual(t0["Command"], { "Ref": "AWS::NoValue" })
assert.equal(t0["Cpu"], "200")
assert.equal(t0["Entrypoint"], null)
assert.deepEqual(t0["Environment"], null)
assert.equal(t0["Essential"], null)
assert.equal(t0["Image"], "docker.io/httpd")
assert.deepEqual(t0["Links"], [])
assert.equal(t0["Memory"], "300")
assert.deepEqual(t0["MountPoints"], null)
assert.equal(t0["Name"], "web")
assert.deepEqual(t0["PortMappings"], [ { "HostPort": 53081, "ContainerPort": 80 } ])

assert.deepEqual(t0["Volumes"], null)
assert.equal(t0["Key"], null)
assert.deepEqual(t0["Services"], null)

assert.equal(t1["Command"], null)
assert.equal(t1["Cpu"], "20")
assert.equal(t1["Entrypoint"], null)
assert.deepEqual(t1["Environment"], [
  { "Name": "AWS_REGION", "Value": "us-east-1" },
  { "Name": "AWS_ACCESS", "Value": { "Ref": "LogsAccess" } },
  { "Name": "AWS_SECRET", "Value": { "Fn::GetAtt": [ "LogsAccess", "SecretAccessKey" ] } },
  { "Name": "KINESIS", "Value": { "Ref": "Kinesis" } },
  { "Name": "CONTAINERS", "Value": "web" },
])
assert.equal(t1["Essential"], null)
assert.equal(t1["Image"], "index.docker.io/convox/logs")
assert.deepEqual(t1["Links"], ["web:web"])
assert.equal(t1["Memory"], "64")
assert.deepEqual(t1["MountPoints"], [ { "ContainerPath": "/var/run/docker.sock", "SourceVolume": "Docker" } ])
assert.equal(t1["Name"], "convox-logs")
assert.deepEqual(t1["PortMappings"], null)

assert.deepEqual(t1["Volumes"], null)
assert.equal(t1["Key"], null)
assert.deepEqual(t1["Services"], null)
