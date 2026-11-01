pipeline {
    agent any

    options {
        timestamps()
        skipDefaultCheckout(true)
    }

    environment {
        REGISTRY = 'ghcr.io/barbodimani81'
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

                    # 8080 is Jenkins itself, so start at 18080. Each executor gets its own port,
                    # so a presubmit and a postsubmit running at once don't fight over one.
                    PORT=$((18080 + EXECUTOR_NUMBER))

                    # CI has no database: point at a port where nothing listens.
                    PORT=$PORT APP_VERSION=ci APP_ENV=ci DATABASE_URL=postgres://ci@127.0.0.1:1/none "$BIN" &
                    PID=$!
                    # Always stop the server, even when a check fails, so it can't hold the port for the next build.
                    trap 'kill $PID' EXIT

                    # Retry while the server starts up instead of failing on the first refused connection.
                    curl -sf --retry 10 --retry-connrefused --retry-delay 1 http://localhost:$PORT/healthz

                    # Without a database the Pod must report not ready, so Kubernetes sends it no traffic.
                    READY=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:$PORT/readyz)
                    if [ "$READY" != "503" ]; then
                        echo "/readyz returned $READY without a database, want 503"
                        exit 1
                    fi

                    echo "Smoke test passed."
                '''
            }
        }
        stage('Render Kustomize') {
            steps {
                sh '''
                    # Catch broken YAML or a bad patch in review, not at deploy time.
                    for overlay in deploy/overlays/*/; do
                        echo "Rendering $overlay"
                        kubectl kustomize "$overlay" > /dev/null
                    done
                '''
            }
        }

        stage('Publish Images') {
            // Only task-postsubmit (main moved) publishes; unreviewed patch sets never push.
            when { environment name: 'GERRIT_EVENT_TYPE', value: 'ref-updated' }
            steps {
                withCredentials([usernamePassword(credentialsId: 'ghcr', usernameVariable: 'GHCR_USER', passwordVariable: 'GHCR_TOKEN')]) {
                    sh '''
                        # One immutable tag per commit, shared by both images. Never :latest.
                        TAG=$(git rev-parse --short=7 HEAD)
                        LABELS="--label org.opencontainers.image.source=https://github.com/barbodimani81/task --label org.opencontainers.image.revision=$TAG"

                        # Throwaway login dir, so the token never lands in Jenkins' home.
                        export DOCKER_CONFIG=$(mktemp -d)
                        trap 'rm -rf "$DOCKER_CONFIG"' EXIT
                        echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin

                        docker build $LABELS --build-arg VERSION="$TAG" -t "$REGISTRY/task-app:$TAG" .
                        docker build $LABELS -t "$REGISTRY/task-migrate:$TAG" migrations

                        docker push "$REGISTRY/task-app:$TAG"
                        docker push "$REGISTRY/task-migrate:$TAG"
                    '''
                }
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
