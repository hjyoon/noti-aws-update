from locust import HttpUser, task, between


class QuickstartUser(HttpUser):
    wait_time = between(0, 1)

    # @task(1)
    # def req_index(self):
    #     self.client.get("/")

    # @task(1)
    # def req_health(self):
    #     self.client.get("/health")

    @task(1)
    def req_tags(self):
        self.client.get("/api/tags")

    @task(1)
    def req_tags_with_search(self):
        self.client.get("/api/tags?name=ec2")

    @task(1)
    def req_whatsnews(self):
        self.client.get("/api/whatsnews")

    @task(1)
    def req_whatsnews_with_tagsid(self):
        self.client.get("/api/whatsnews?tags=2282")

    @task(1)
    def req_whatsnews_with_search(self):
        self.client.get("/api/whatsnews?search=ec2")

    @task(1)
    def req_whatsnews_with_tagsid_and_search(self):
        self.client.get("/api/whatsnews?tags=2282&search=ec2")
