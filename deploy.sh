#!/bin/sh

TAG="v0.0.9" \
&& aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 767398145752.dkr.ecr.us-west-2.amazonaws.com \
&& docker build -t 767398145752.dkr.ecr.us-west-2.amazonaws.com/bp-tracker:$TAG . \
&& docker push 767398145752.dkr.ecr.us-west-2.amazonaws.com/bp-tracker:$TAG \
&& cd terraform \
&& terraform apply -var ecr_image_tag=$TAG \
&& cd ..
