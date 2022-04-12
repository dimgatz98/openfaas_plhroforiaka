import os, json, sys
from pymongo import MongoClient
from urllib.parse import quote_plus

def get_uri():
    password=""
    with open("/var/openfaas/secrets/mongo-db-password") as f:
        password = f.read()

    return "mongodb://%s:%s@%s" % (
    quote_plus("root"), quote_plus(password), os.getenv("mongo_host"))

def handle(req):
    """handle a request to the function
    Args:
        req (str): request body
    """

    method = os.getenv("Http_Method")
    sys.stderr.write("Method: {}\n".format(method))

    if method in ["POST", "PUT"]:
        uri = get_uri()
        client = MongoClient(uri)

        db = client['openfaas']
        followers = db.followers
        for follower in followers.find():
            if req.strip() == follower[u'username']:
                return "A user with that username already exists"

        follower={"username": req.strip()}
        res = followers.insert_one(follower)
        return "Record inserted: {}".format(res.inserted_id)

    elif method == "GET":
        uri = get_uri()
        client = MongoClient(uri)

        db = client['openfaas']
        followers = db.followers

        ret = []
        for follower in followers.find():
            ret.append({"username": follower[u'username']})

        return json.dumps(ret)

    elif method == "DELETE":
        uri = get_uri()
        client = MongoClient(uri)

        db = client['openfaas']
        followers = db.followers

        if req.strip() == "all":
            followers.drop()
            return "Collection deleted"

        for follower in followers.find():
            if follower[u'username'] == req.strip():
                found = True
                break    

        found = False
        for follower in followers.find():
            if follower[u'username'] == req.strip():
                found = True
                break

        if not found:
            return "No users with username {}".format(req.strip())

        follower = {"username": req.strip()}
        res = followers.delete_one(follower)
        return "Record deleted: {}".format(req.strip() )


    return "Method: {} not supported".format(method)
