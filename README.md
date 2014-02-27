# s3-user-script

This is a simple tool for bootstrapping EC2 instances from a
`user-script` in a private S3 bucket.

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

3. Upload your `user-script` to `s3://my-user-scripts/MyRole/user-script`.

4. Create your instance with the following `user-script`:

   ```bash
   #!/bin/bash

   # fail fast
   set -o errexit
   set -o nounset
   set -o pipefail

   curl -fLOsS https://github.com/grosskur/s3-user-script/releases/download/v20140226/s3-user-script
   chmod 755 s3-user-script
   exec ./s3-user-script -b my-user-scripts
   ```

Congratulations! Your EC2 instance will now run your `user-script`
from S3 on boot.
