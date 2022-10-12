# Configuring OAuth

## Microsoft

https://docs.microsoft.com/en-us/exchange/client-developer/exchange-web-services/how-to-authenticate-an-ews-application-by-using-oauth

https://docs.microsoft.com/en-us/exchange/client-developer/legacy-protocols/how-to-authenticate-an-imap-pop-smtp-application-by-using-oauth

### Prerequisites

- Microsoft Office 365 Administrator account
- Able administrator credentials
- Experience managing Exchange and mailbox permissions

### Set up the App Registration in Azure

From the Azure Active Directory admin center, complete the following:

- Click All services.
- Select App Registrations.
- Click New Registration.

### Allow IMAP4 for Exchange

```ps
Install-Module -Name ExchangeOnlineManagement
Import-module ExchangeOnlineManagement
Connect-ExchangeOnline -Organization Ableai
New-ServicePrincipal -Organization Ableai -AppId {xxxx} -ServiceId {xxxx}
Get-ServicePrincipal -Organization Ableai | fl
```

Now your Exchange is allowed for IMAP with OAuth

To allow use login with OAuth

```ps
Add-MailboxPermission -Identity denis@ekspand.onmicrosoft.com -User denis@ekspand.onmicrosoft.com -AccessRights FullAccess
```

## Google

https://developers.google.com/gmail/imap/xoauth2-protocol

