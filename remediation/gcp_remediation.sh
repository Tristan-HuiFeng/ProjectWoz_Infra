#!/bin/bash

project="the-other-450607-a4"
echo "ProjectId: $project"

for bucket in $(gcloud storage ls --project $project); do
  echo "    -> Bucket $bucket"

  $(gcloud storage buckets update $bucket --public-access-prevention)  
  $(gcloud storage buckets update $bucket --soft-delete-duration=7d)
  $(gcloud storage buckets update $bucket ----retention-period=7d)
  $(gcloud storage buckets update $bucket --uniform-bucket-level-access)

done