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

        stage('Smoke Test') {
            steps {
                sh '''
                    bazel build //cmd/server:server
                    BIN=$(bazel cquery --output=files //cmd/server:server 2>/dev/null)

                    # 8080 is Jenkins itself inside this container, so use 18080.
                    # CI has no database: point at a port where nothing listens.
                    PORT=18080 APP_VERSION=ci APP_ENV=ci DATABASE_URL=postgres://ci@127.0.0.1:1/none "$BIN" &
                    PID=$!
                    # Always stop the server, even when a check fails, so it can't hold the port for the next build.
                    trap 'kill $PID' EXIT

                    # Retry while the server starts up instead of failing on the first refused connection.
                    curl -sf --retry 10 --retry-connrefused --retry-delay 1 http://localhost:18080/healthz

                    # Without a database the Pod must report not ready, so Kubernetes sends it no traffic.
                    READY=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:18080/readyz)
                    if [ "$READY" != "503" ]; then
                        echo "/readyz returned $READY without a database, want 503"
                        exit 1
                    fi

                    echo "Smoke test passed."
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
