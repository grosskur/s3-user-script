# s3-user-script

Securely bootstrap an [EC2][ec2] instance using [IAM Role][iam-role]
credentials to download and run a [User Data][user-data] script from a
private [S3][s3] bucket.

## Getting started

### Precompiled binaries

* [Linux (64-bit)](https://github.com/grosskur/s3-user-script/releases/download/v20140226/s3-user-script)

### Compile from source

```bash
$ go get -u github.com/grosskur/s3-user-script
```

## Background

The EC2 platform provides several features that can be used together
to create elastic, dynamically configured clusters of machines:

* [User Data][user-data] is arbitrary data that you can provide when
  you launch an instance. If this data is a shell script, it will be
  executed the first time the instance is booted.

* A [Launch Configuration][launch-configuration] is a template for
  launching repeated instances with the same parameters. It can also
  have User Data associated with it, which is passed on to each
  instance you launch.

* An [Auto-Scaling Group][auto-scaling-group] ties together a Launch
  Configuration to a Scaling Plan to let you dynamically grow or
  shrink a group of instances.

* An [IAM Role][iam-role] can be assigned to an instance or launch
  configuration to generate a temporary, automatically-rotated set of
  AWS credentials for that particular instance.

### Problem

User scripts work fine when launching a single instance. However, when
used with an auto-scaling group, you are essentially "baking" the data
up-front into all the instances you will launch. The only way to
change the user data is to destroy and recreate the launch
configuration associated with the auto-scaling group.

### Solution

`s3-user-script` is a shim that simply downloads the real user script
from an S3 bucket and runs it. Since the S3 bucket should be private,
IAM role credentials are used to access it. And to keep things simple,
it assumes your user scripts are organized based on the role name
(although this is configurable).

## Usage

1. Create an S3 bucket `my-user-scripts`.

2. Create an IAM role `MyRole`. Give it access to your bucket with a
   policy like the following:

   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": "s3:*",
         "Resource": [
           "arn:aws:s3:::my-user-scripts",
           "arn:aws:s3:::my-user-scripts/*",
         ]
       }
     ]
   }
   ```

3. Create a `user-script` and upload it to `s3://my-user-scripts/MyRole/user-script`.

4. Create your instance with the following user data:

   ```bash
   #!/bin/bash -e
   curl -fLOsS https://github.com/grosskur/s3-user-script/releases/download/v20140226/s3-user-script
   chmod 755 s3-user-script
   exec ./s3-user-script -b my-user-scripts
   ```

   Alternatively, if you bake `/usr/local/bin/s3-user-script` into
   your AMI (using a tool like [Packer][packer]), your user data
   becomes even simpler:

   ```bash
   #!/bin/bash -e
   exec s3-user-script -b my-user-scripts
   ```

Congratulations! Your EC2 instances will now run the latest version of
your role-specific user scripts on boot. Changes to the user scripts
go live immediately when you update them on S3.

[ec2]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/concepts.html
[iam-role]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html
[user-data]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html
[s3]: http://docs.aws.amazon.com/AmazonS3/latest/dev/Welcome.html
[launch-configuration]: http://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/WorkingWithLaunchConfig.html
[auto-scaling-group]: http://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/WorkingWithASG.html
[scaling-plan]: http://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/scaling_plan.html
[packer]: https://github.com/mitchellh/packer
