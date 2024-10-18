package utils

// Stored in redis as a hash set with the key being client-id and the value being the client-secret
var ClientCredentialsRedis = "client-credentials"

// Format stored in redis is access-tokens:{token} as the key and the value is the client secret
// Structure is a regular key-value pair
var AccessTokenRedis = "access-tokens"
