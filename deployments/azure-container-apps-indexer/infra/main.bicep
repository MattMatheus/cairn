targetScope = 'resourceGroup'

@description('Azure region for the Container Apps resources.')
param location string = resourceGroup().location

@description('Short environment name used in generated resource names.')
param environmentName string

@description('Container image for the Cairn indexer service.')
param containerImage string

@description('Cairn workspace id served by this indexer.')
param workspaceId string

@description('Existing Azure Storage account name that contains workspace blobs.')
param storageAccountName string

@description('Blob container containing workspace documents.')
param storageContainerName string

@description('Optional prefix within the storage container.')
param storagePrefix string = ''

@description('Expected bearer token audience or app id URI.')
param authAudience string

@description('Expected Microsoft Entra tenant id.')
param authTenantId string

@description('Name of the ACA secret that operators create for the PostgreSQL DSN.')
param postgresDsnSecretName string = 'postgres-dsn'

@description('Minimum replica count. Keep above zero when cold starts are unacceptable.')
param minReplicas int = 0

@description('Maximum replica count.')
param maxReplicas int = 3

var namePrefix = 'cairn-${environmentName}'
var logAnalyticsName = '${namePrefix}-logs'
var managedEnvironmentName = '${namePrefix}-aca'
var identityName = '${namePrefix}-indexer-mi'
var containerAppName = '${namePrefix}-indexer'

resource logs 'Microsoft.OperationalInsights/workspaces@2023-09-01' = {
  name: logAnalyticsName
  location: location
  properties: {
    sku: {
      name: 'PerGB2018'
    }
    retentionInDays: 30
  }
}

resource identity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: identityName
  location: location
}

resource storage 'Microsoft.Storage/storageAccounts@2023-01-01' existing = {
  name: storageAccountName
}

resource environment 'Microsoft.App/managedEnvironments@2024-03-01' = {
  name: managedEnvironmentName
  location: location
  properties: {
    appLogsConfiguration: {
      destination: 'log-analytics'
      logAnalyticsConfiguration: {
        customerId: logs.properties.customerId
        sharedKey: logs.listKeys().primarySharedKey
      }
    }
  }
}

resource storageReader 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storage.id, identity.id, 'Storage Blob Data Reader')
  scope: storage
  properties: {
    principalId: identity.properties.principalId
    principalType: 'ServicePrincipal'
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '2a2b9908-6ea1-4ae2-8e65-a410df84e7d1')
  }
}

resource app 'Microsoft.App/containerApps@2024-03-01' = {
  name: containerAppName
  location: location
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${identity.id}': {}
    }
  }
  properties: {
    managedEnvironmentId: environment.id
    configuration: {
      activeRevisionsMode: 'Single'
      ingress: {
        external: true
        targetPort: 8080
        transport: 'http'
        allowInsecure: false
      }
    }
    template: {
      containers: [
        {
          name: 'indexer'
          image: containerImage
          env: [
            {
              name: 'CAIRN_INDEXER_WORKSPACE_ID'
              value: workspaceId
            }
            {
              name: 'CAIRN_STORAGE_ACCOUNT'
              value: storageAccountName
            }
            {
              name: 'CAIRN_STORAGE_CONTAINER'
              value: storageContainerName
            }
            {
              name: 'CAIRN_STORAGE_PREFIX'
              value: storagePrefix
            }
            {
              name: 'CAIRN_POSTGRES_DSN_SECRET_NAME'
              value: postgresDsnSecretName
            }
            {
              name: 'CAIRN_AUTH_AUDIENCE'
              value: authAudience
            }
            {
              name: 'CAIRN_AUTH_TENANT_ID'
              value: authTenantId
            }
            {
              name: 'CAIRN_LOG_LEVEL'
              value: 'info'
            }
          ]
          resources: {
            cpu: json('0.5')
            memory: '1Gi'
          }
        }
      ]
      scale: {
        minReplicas: minReplicas
        maxReplicas: maxReplicas
      }
    }
  }
}

output containerAppName string = app.name
output containerAppFqdn string = app.properties.configuration.ingress.fqdn
output managedIdentityClientId string = identity.properties.clientId
