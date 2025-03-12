module "lambda" {
  source = "./modules/lambda"
  discovery_sqs_queue_arn = module.sqs.discovery_sqs_queue_arn
  retrieval_sqs_queue_arn = module.sqs.retrieval_sqs_queue_arn
}

module "sqs" {
  source = "./modules/sqs"
}

# module "rds" {
#   source        = "./modules/rds"
#   subnet_db_ids = module.vpc.subnet_db_id
#   vpc_id        = module.vpc.vpc_id
# }

# module "s3" {
#   source = "./modules/s3"
# }

# module "rabbitmq" {
#   source            = "./modules/mq-broker"
#   rabbitmq_username = var.rabbitmq_username
#   rabbitmq_password = var.rabbitmq_password
# }

# # module "ecs_ec2" {
# #   source = "./modules/ecs-ec2"
# #   vpc_id = module.vpc.vpc_id
# # }

# module "security_groups" {
#   source = "./modules/security-group"
#   vpc_id = module.vpc.vpc_id
# }

# module "alb" {
#   source         = "./modules/alb"
#   vpc_id         = module.vpc.vpc_id
#   subnet_web_ids = module.vpc.subnet_web_id
#   alb_sg_id      = module.security_groups.alb_sg_id
# }

# module "cloud_map" {
#   source = "./modules/cloud-map"
#   vpc_id = module.vpc.vpc_id
# }

# module "ecs" {
#   source = "./modules/ecs-fargate"
#   vpc_id = module.vpc.vpc_id
#   # alb_app_tg_arn = module.alb.alb_app_tg_arn
#   target_group_arns                  = module.alb.target_group_arns
#   subnet_app_ids                     = module.vpc.subnet_app_id
#   ecs_sg_id                          = module.security_groups.ecs_sg_id
#   property_service_discovery_arn     = module.cloud_map.property_discovery_service_arn
#   availability_service_discovery_arn = module.cloud_map.availability_discovery_service_arn
#   user_service_discovery_arn         = module.cloud_map.user_discovery_service_arn
#   booking_service_discovery_arn      = module.cloud_map.booking_discovery_service_arn
#   payment_service_discovery_arn      = module.cloud_map.payment_discovery_service_arn

#   depends_on = [module.alb, module.cloud_map]
# }
