ACCOUNT="123456789012"

tee -a ~/.docker/config.json <<EOF
{
	"credHelpers": {
		"public.ecr.aws": "ecr-login",
		"$ACCOUNT.dkr.ecr.<region>.amazonaws.com": "ecr-login"
	}
}
EOF

sudo apt update
sudo apt install amazon-ecr-credential-helper
docker pull $ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/cosmostesting:validatorrewards-1.0
