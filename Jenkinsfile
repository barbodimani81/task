pipeline {
    agent any

    options {
        timestamps()
        skipDefaultCheckout(true)
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Check Go Formatting') {
            steps {
                sh '''
                    UNFORMATTED=$(gofmt -l $(find . -name '*.go' -not -path './vendor/*'))

                    if [ -n "$UNFORMATTED" ]; then
                        echo "The following Go files need formatting:"
                        echo "$UNFORMATTED"
                        exit 1
                    fi

                    echo "All Go files are formatted."
                '''
            }
        }

        stage('Go Tests') {
            steps {
                sh 'go test ./...'
            }
        }

        stage('Bazel Build') {
            steps {
                sh 'bazel build //...'
            }
        }

        stage('Bazel Tests') {
            steps {
                sh 'bazel test //...'
            }
        }

        stage('Run Application') {
            steps {
                sh '''
                    OUTPUT=$(go run ./cmd/server)
                    echo "$OUTPUT"

                    if [ "$OUTPUT" != "Hello, World!" ]; then
                        echo "Unexpected application output."
                        exit 1
                    fi
                '''
            }
        }
    }

    post {
        success {
            echo 'CI pipeline completed successfully.'
        }

        failure {
            echo 'CI pipeline failed. Check the stage logs.'
        }

        always {
            echo 'Pipeline finished.'
        }
    }
}
