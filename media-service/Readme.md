```
# Video Related APIs 
getAllVideo: Get Catalog Metadata
Endpoint: GET /catalog/titles
Description: Retrieves metadata for all titles in the catalog.
Parameters:
region (string): The region for which to retrieve the catalog.
language (string): The language in which to retrieve metadata.

getVideo: Get Video Details
Endpoint: GET /catalog/titles/{titleId}
Description: Retrieves detailed information about a specific title.
Parameters:
titleId (string): The unique ID of the title.

searchVideo: Search for video based on title or tag (query, nextpage). 
EndPoint: GET /catalog/titles/{tag}
Parameter: None
return [][]bytes.

Streaming Video - GET (Video id, codec, resolution, offset). 
return Video stream (Byte stream).


# Netflix User API
Used to manage user accounts and preferences.

getuser: Get User Profile
Endpoint: GET /users/{userId}/profile
Description: Retrieves the profile information of a user.
Parameters:
userId (string): The unique ID of the user.

setUserPref: Update User Preferences
Endpoint: PUT /users/{userId}/preferences
Description: Updates the preferences of a user.
Parameters:
userId (string): The unique ID of the user.
preferences (JSON): The preferences to update.

getRecommendation: Get Recommendations
Endpoint: GET /users/{userId}/recommendations
Description: Retrieves content recommendations for a user.
Parameters:
userId (string): The unique ID of the user.
limit (integer): The number of recommendations to retrieve.

playback: Start Playback
Endpoint: /playback/start
Description: Starts playback of a specific title.
Method: POST
Parameters:
titleId (string): The unique ID of the title.
userId (string): The unique ID of the user.
```