package gen 

const (
RUBY = `FROM public.ecr.aws/lambda/ruby:2.7

COPY app.rb ./

# Command can be overwritten by providing a different command in the template directly.
CMD ["app.lambda_handler"]`)
