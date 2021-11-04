# playlist-protector

Playlist protector is a Golang application that continuously checks and reinstates a Spotify playlist's name when it's taken down by bots (or people) that falsely report it as offensive.

## Usage

The service is intended to be deployed to a Kubernetes cluster using the manifests found in [./kube](./kube). You will need to create an app from the [Spotify developer portal](https://developer.spotify.com/), configuring the app with a redirect URI that will be able to reach the service in your cluster. Make sure to note down your app's client ID and client secret.

Once you've setup a Spotify app, the service can be configured and deployed with the following settings:

`address` - The IP address to run the OAuth callback server on, defaults to `0.0.0.0`  
`port` - The port to run the OAuth callback server on, defaults to `8080`  
`playlist-id` - The ID of the Spotify playlist to monitor  
`playlist-name` - The name to reinstate should the playlist's name be changed or removed  
`client-id` - The client ID of the Spotify app created from the developer portal  
`client-secret` - The client secret of the Spotify app created from the developer portal  
`redirect-uri` - The redirect URI, used to reach the service running in your cluster (must match the redirect URI in app settings)  
`debug` - Can be set to enable debug logging  

When the service initially starts, it will print a URL to STDOUT which you must manually visit to start authorisation using OAuth. This only needs to happen once at startup, as the service will automatically refresh the token it receives at the end of the OAuth flow.

## Considerations

This service is currently experimental and is intended to be used by developers/cluster operators. With some future additions it would be possible to host this as a proper Spotify app capable of handling multiple users and playlists, with a nice landing page for onboarding and kicking off the OAuth flow.