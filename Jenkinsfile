pipeline {
    agent any
    environment {
        IMAGE_NAME = "agent"
        JOB_NAME = "ServerRun Docker Agent"
        CODING_DOCKER_REG_HOST = "${env.CCI_CURRENT_TEAM}-docker.pkg.${env.CCI_CURRENT_DOMAIN}"
        DOCKER_IMAGE_NAME = "${env.CODING_DOCKER_REG_HOST}/${env.PROJECT_NAME}/images/${IMAGE_NAME}"
    }
    stages {
        stage('检出') {
            steps {
                checkout([$class: 'GitSCM', branches: [[name: env.GIT_BUILD_REF]],
                userRemoteConfigs: [[url: env.GIT_REPO_URL, credentialsId: env.CREDENTIALS_ID]]])
            }
        }
        stage('构建镜像并推送') {
            steps {
                script {
                    docker.withRegistry("https://${env.CODING_DOCKER_REG_HOST}", "${env.CODING_ARTIFACTS_CREDENTIALS_ID}") {
                        def img = docker.build("${env.DOCKER_IMAGE_NAME}:latest")
                        img.push()
                    }
                }
            }
        }
    }
}