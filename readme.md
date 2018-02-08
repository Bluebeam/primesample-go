# Studio Prime Sample - Go

## Resources

- [API Documentation](https://studioapi.bluebeam.com/publicapi/swagger/ui/index)

## Workflow

The sample application demonstrates the following scenario:

1. Authorizes the app using Three-Legged OAuth/2
2. Upload a file to a Studio Project and then checks it out to a new Studio Session
    * User Chooses a Project from a drop-down
    * User Specifies a Session Name
    * User browses for a File
    * User Clicks Create
3. The back-end application now completes the following steps
    * Starts an upload to the Project which gets an AWS Upload URL
    * Uploads the file to AWS
    * Confirms the Upload in the Project
    * Creates a new Session
    * Checks out the file to the Session
4. Users adds markups to the file while it is in a Session
5. User clicks 'Finish' button in application which then does the following
    * Sets the Session state to 'Finalizing' to kick everyone out of the Session
    * Kicks off a process to generate a snapshot of the file with the markups
    * Waits for the snapshot to finish
    * Downloads the snapshot
    * Deletes the Session
    * Starts a checkin for the project file, getting an AWS Upload URL
    * Uploads the file to AWS
    * Confirms the project Checkin
    * Kicks off a job to flatten the file
    * Gets a share link for the project file

## Notes

In Step 5 above, an alternate approach could have been chosen. Instead of generating the snapshot, the file could have been checked in directly from the Session. However, the call to check in from Session only updates the project copy of the file leaving the file remaining in a checked out state. Furthermore, there is no convenient way to poll the status of the checkin. In this scenario the application would either have to poll the chat history looking for a specific chat message that identifies when the update is complete, or the application could poll the file history for an updated revision. Once the update is complete, the checkout could be undone on the file and the Session deleted as in the above workflow.

For the purpose of this sample the snapshot method is more straightforward with the caveat the extra bandwidth is consumed by downloading the file and then re-uploading it.

## Details

### Configuration

Configuration of the client id, secret, and callback urls are handled by using either a config file or using environment variables. To use a config file, place the following as config.json in the root of the project:

```
{
    "clientId": "CLIENT_ID_GOES_HERE",
    "clientSecret": "CLIENT_SECRET_GOES_HERE", 
    "url": "http://localhost:5000",
}
```

The secret is used for encrypting the session cookie.

If environment variables are used, they are:

- CLIENT_ID
- CLIENT_SECRET
- URL

### Authentication

The app uses the standard oauth2 at golang.org/x/oauth2. Some extra code was developed to be able to intercept a message as to when a token is refreshed so that the app has the opportunity to store a new refresh token. Because Studio Refresh Tokens are one time use only it is imperative that the new ones are saved. This code is found in studiotoken.go. 

### Database

Persistent storage is necessary in order to save tokens. BoltDB is a very simple, pure Go, local database so it was chosen for this sample. As a production system, an external database would be required.