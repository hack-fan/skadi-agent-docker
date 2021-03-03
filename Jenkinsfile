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
                sh """
                    curl '${env.E_WECHAT_ROBOT_WEBHOOK_URL}' \
                         -H 'Content-Type: application/json' \
                         -d '
                         {
                              "msgtype": "markdown",
                              "markdown": {
                                  "content": "[${env.JOB_NAME}:${CI_BUILD_NUMBER}] <font color=\u0027comment\u0027>开始构建</font>"
                              }
                         }'
                """
                checkout([$class: 'GitSCM', branches: [[name: env.GIT_BUILD_REF]],
                userRemoteConfigs: [[url: env.GIT_REPO_URL, credentialsId: env.CREDENTIALS_ID]]])
            }
        }
        stage('构建镜像并推送') {
            steps {
                script {
                    docker.withRegistry("https://${env.CODING_DOCKER_REG_HOST}", "${env.CODING_ARTIFACTS_CREDENTIALS_ID}") {
                        def img = docker.build("${env.DOCKER_IMAGE_NAME}:latest","-f ./cmd/api/Dockerfile .")
                        img.push()
                    }
                }
            }
        }
    }
    post {
        success {
            sh """
                curl '${env.WEWORK_WEBHOOK}' \
                     -H 'Content-Type: application/json' \
                     -d '
                     {
                          "msgtype": "markdown",
                          "markdown": {
                              "content": "[${env.JOB_NAME}:${CI_BUILD_NUMBER}] <font color=\u0027info\u0027>构建成功</font>"
                          }
                     }'
            """
        }
        failure {
            sh """
                curl '${env.WEWORK_WEBHOOK}' \
                     -H 'Content-Type: application/json' \
                     -d '
                     {
                          "msgtype": "markdown",
                          "markdown": {
                              "content": "[${env.JOB_NAME}:${CI_BUILD_NUMBER}] <font color=\u0027warning\u0027>构建失败</font>"
                          }
                     }'
            """
        }
    }
}