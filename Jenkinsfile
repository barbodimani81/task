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

                        # GHCR sometimes refuses the manifest after a slow upload. The layers are
                        # already there by then, so a second try succeeds.
                        push_with_retry() {
                            for attempt in 1 2 3; do
                                docker push "$1" && return 0
                                echo "Push of $1 failed (attempt $attempt of 3), retrying."
                                sleep $((attempt * 10))
                            done
                            return 1
                        }
                        push_with_retry "$REGISTRY/task-app:$TAG"
                        push_with_retry "$REGISTRY/task-migrate:$TAG"
                    '''
                }
            }
        }
        stage('Promote to Dev') {
            // Every merge on main goes to dev by itself. Prod moves only through a change a human approves.
            when { environment name: 'GERRIT_EVENT_TYPE', value: 'ref-updated' }
            steps {
                // A retry starts over from a fresh clone, so it also covers main moving under us.
                retry(3) {
                    sh '''
                        TAG=$(git rev-parse --short=7 HEAD)
                        CONFIG=$(mktemp -d)
                        trap 'rm -rf "$CONFIG"' EXIT

                        git clone -q ssh://jenkins@gerrit:29418/task-config "$CONFIG"
                        curl -fsSLo "$CONFIG/.git/hooks/commit-msg" http://gerrit:8080/tools/hooks/commit-msg
                        chmod +x "$CONFIG/.git/hooks/commit-msg"

                        DEV="$CONFIG/overlays/dev/kustomization.yaml"
                        CURRENT=$(awk '/newTag:/ {print $2; exit}' "$DEV")
                        if [ "$CURRENT" = "$TAG" ]; then
                            echo "Dev already runs $TAG."
                            exit 0
                        fi
                        # Two builds can finish out of order. Never move dev back to an older commit.
                        if ! git merge-base --is-ancestor "$CURRENT" "$TAG" 2>/dev/null; then
                            echo "Dev runs $CURRENT, which is not older than $TAG. Leaving it."
                            exit 0
                        fi

                        sed -i "s/newTag: .*/newTag: $TAG/" "$DEV"
                        kubectl kustomize "$CONFIG/overlays/dev" > /dev/null

                        cd "$CONFIG"
                        git -c user.name='Jenkins CI' -c user.email='jenkins@ci.local' \
                            commit -qam "dev: run $TAG" -m "Automatic: task commit $TAG landed on main."
                        git push origin HEAD:refs/for/main
                        # A normal submit, so Gerrit checks every submit requirement: a bump that touched
                        # prod would stop at Prod-Approval. (Submitting during the push would skip them all.)
                        COMMIT=$(git rev-parse HEAD)
                        if ! ssh -p 29418 jenkins@gerrit gerrit review --project task-config \
                                --code-review +2 --verified +1 --submit "$COMMIT"; then
                            # Don't leave a stale change behind. The retry starts over with a new one.
                            ssh -p 29418 jenkins@gerrit gerrit review --project task-config --abandon "$COMMIT"
                            exit 1
                        fi
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
