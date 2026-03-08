param(
  [string]$BaseUrl = "http://127.0.0.1:7788",
  [string]$Login = "dief",
  [string]$Password = "PainelDief#2026",
  [string]$OpsToken = "",
  [switch]$MutatingChecks
)

$ErrorActionPreference = "Stop"

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$result = [ordered]@{}

$health = Invoke-RestMethod -Uri "$BaseUrl/api/health" -WebSession $session -UseBasicParsing
$result.health = [string]$health.status
$result.panelOnly = [bool]$health.panelOnly

$ready = Invoke-RestMethod -Uri "$BaseUrl/api/ready" -WebSession $session -UseBasicParsing
$result.ready = [bool]$ready.ready

$root = Invoke-WebRequest -Uri "$BaseUrl/" -WebSession $session -UseBasicParsing
$result.root = [int]$root.StatusCode

$loginBody = @{
  login = $Login
  password = $Password
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/login" -Method Post -WebSession $session -ContentType "application/json" -Body $loginBody
$result.login = [bool]$loginResp.ok

$bootstrap = $loginResp.bootstrap
$result.rooms = @($bootstrap.rooms).Count
$result.viewer = [string]$bootstrap.viewer.displayName
$room = @($bootstrap.rooms) | Where-Object { $_.slug -eq "chat-geral" } | Select-Object -First 1
if (-not $room) {
  throw "chat-geral nao encontrada no bootstrap"
}

$profile = Invoke-RestMethod -Uri "$BaseUrl/api/panel/social/profile?userId=$($bootstrap.viewer.id)" -WebSession $session -UseBasicParsing
$result.profile = [bool]($profile.profile.user.userId -eq $bootstrap.viewer.id)

$messages = Invoke-RestMethod -Uri "$BaseUrl/api/panel/messages?roomId=$($room.id)&limit=8" -WebSession $session -UseBasicParsing
$result.messagesListed = @($messages.messages).Count

$pollList = Invoke-RestMethod -Uri "$BaseUrl/api/panel/polls?roomId=$($room.id)&limit=6" -WebSession $session -UseBasicParsing
$result.pollsListed = @($pollList.polls).Count

$search = Invoke-RestMethod -Uri "$BaseUrl/api/panel/search?query=dief&limit=5" -WebSession $session -UseBasicParsing
$result.search = ($null -ne $search.messages)

$legacyStatus = 0
try {
  $legacyResp = Invoke-WebRequest -Uri "$BaseUrl/api/plugins" -WebSession $session -UseBasicParsing
  $legacyStatus = [int]$legacyResp.StatusCode
} catch {
  if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
    $legacyStatus = [int]$_.Exception.Response.StatusCode.value__
  } else {
    throw
  }
}
$result.legacyClosed = ($legacyStatus -eq 404)

$opsHeaders = @{}
if ($OpsToken) {
  $opsHeaders["Authorization"] = "Bearer $OpsToken"
}
if ($BaseUrl -like "http://127.0.0.1:*" -or $BaseUrl -like "http://localhost:*" -or $OpsToken) {
  $ops = Invoke-RestMethod -Uri "$BaseUrl/api/ops/summary" -WebSession $session -Headers $opsHeaders -UseBasicParsing
  $result.opsUsers = [int]$ops.summary.users
  $result.opsRooms = [int]$ops.summary.rooms
  $result.opsMessages = [int]$ops.summary.messages
}

if ($MutatingChecks) {
  $stamp = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
  $originalViewer = $bootstrap.viewer

  $messageResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/messages" -Method Post -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    body = "smoke-message-$stamp"
    kind = "text"
    attachment = $null
    replyToId = 0
  } | ConvertTo-Json)
  $result.messagePost = [bool]($messageResp.message.id -gt 0)

  $messageId = [int64]$messageResp.message.id

  $editResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/messages" -Method Put -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    messageId = $messageId
    body = "smoke-edit-$stamp"
  } | ConvertTo-Json)
  $result.messageEdit = [bool]($editResp.message.body -eq "smoke-edit-$stamp")

  $reactionResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/reactions/toggle" -Method Post -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    messageId = $messageId
    emoji = ":fire:"
  } | ConvertTo-Json)
  $result.reaction = [bool](@($reactionResp.message.reactions).Count -gt 0)

  $favoriteResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/favorites/toggle" -Method Post -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    messageId = $messageId
  } | ConvertTo-Json)
  $result.favorite = [bool]$favoriteResp.favorited

  $pinResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/pins/toggle" -Method Post -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    messageId = $messageId
  } | ConvertTo-Json)
  $result.pin = [bool]$pinResp.pinned

  $deleteResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/messages" -Method Delete -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    messageId = $messageId
  } | ConvertTo-Json)
  $result.messageDelete = [bool]($deleteResp.messageId -eq $messageId)

  $pollResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/polls" -Method Post -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    question = "Smoke // qual modo? $stamp"
    options = @("casual", "ranked")
  } | ConvertTo-Json)
  $result.poll = [bool]($pollResp.poll.id -gt 0)
  $pollDeleteResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/polls" -Method Delete -WebSession $session -ContentType "application/json" -Body (@{
    roomId = [int64]$room.id
    pollId = [int64]$pollResp.poll.id
  } | ConvertTo-Json)
  $result.pollDelete = [bool]($pollDeleteResp.pollId -eq [int64]$pollResp.poll.id)

  $eventResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/events" -Method Post -WebSession $session -ContentType "application/json" -Body (@{
    title = "Smoke Event $stamp"
    description = "rodada controlada de validacao"
    roomId = [int64]$room.id
    startsAt = (Get-Date).ToUniversalTime().AddHours(2).ToString("o")
  } | ConvertTo-Json)
  $result.event = [bool]($eventResp.event.id -gt 0)
  $eventDeleteResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/events" -Method Delete -WebSession $session -ContentType "application/json" -Body (@{
    eventId = [int64]$eventResp.event.id
  } | ConvertTo-Json)
  $result.eventDelete = [bool]($eventDeleteResp.eventId -eq [int64]$eventResp.event.id)

  $profileStamp = "smoke-status-$stamp"
  Invoke-RestMethod -Uri "$BaseUrl/api/panel/profile" -Method Post -WebSession $session -ContentType "application/json" -Body (@{
    displayName = [string]$originalViewer.displayName
    bio = [string]$originalViewer.bio
    theme = [string]$originalViewer.theme
    accentColor = [string]$originalViewer.accentColor
    avatarUrl = [string]$originalViewer.avatarUrl
    status = [string]$originalViewer.status
    statusText = $profileStamp
  } | ConvertTo-Json) | Out-Null

  $session2 = New-Object Microsoft.PowerShell.Commands.WebRequestSession
  $loginResp2 = Invoke-RestMethod -Uri "$BaseUrl/api/panel/login" -Method Post -WebSession $session2 -ContentType "application/json" -Body $loginBody
  $result.profilePersist = [bool]($loginResp2.bootstrap.viewer.statusText -eq $profileStamp)

  $userName = "smoke" + $stamp
  $createUserResp = Invoke-RestMethod -Uri "$BaseUrl/api/panel/users" -Method Post -WebSession $session2 -ContentType "application/json" -Body (@{
    username = $userName
    displayName = "Smoke " + $stamp
    email = "$userName@local.test"
    password = "Senha#123456"
    role = "member"
  } | ConvertTo-Json)

  $session3 = New-Object Microsoft.PowerShell.Commands.WebRequestSession
  Invoke-RestMethod -Uri "$BaseUrl/api/panel/login" -Method Post -WebSession $session3 -ContentType "application/json" -Body $loginBody | Out-Null
  $profileCheck = Invoke-RestMethod -Uri "$BaseUrl/api/panel/social/profile?userId=$($createUserResp.user.id)" -WebSession $session3 -UseBasicParsing
  $result.userPersist = [bool]($profileCheck.profile.user.userId -eq $createUserResp.user.id)

  Invoke-RestMethod -Uri "$BaseUrl/api/panel/profile" -Method Post -WebSession $session3 -ContentType "application/json" -Body (@{
    displayName = [string]$originalViewer.displayName
    bio = [string]$originalViewer.bio
    theme = [string]$originalViewer.theme
    accentColor = [string]$originalViewer.accentColor
    avatarUrl = [string]$originalViewer.avatarUrl
    status = [string]$originalViewer.status
    statusText = [string]$originalViewer.statusText
  } | ConvertTo-Json) | Out-Null
}

$result.GetEnumerator() | ForEach-Object {
  "{0}={1}" -f $_.Key, $_.Value
}
