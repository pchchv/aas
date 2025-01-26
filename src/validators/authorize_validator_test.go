package validators

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pchchv/aas/src/constants"
	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database/mocks"
	"github.com/pchchv/aas/src/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateScopes(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	tests := []struct {
		name          string
		scope         string
		mockSetup     func()
		expectedError string
	}{
		{
			name:          "Empty scope",
			scope:         "",
			expectedError: "The 'scope' parameter is missing. Ensure to include one or more scopes, separated by spaces. Scopes can be an OpenID Connect scope, a resource:permission scope, or a combination of both.",
		},
		{
			name:          "Valid OpenID Connect scope",
			scope:         "openid profile email",
			expectedError: "",
		},
		{
			name:          "Valid offline_access scope",
			scope:         "offline_access",
			expectedError: "",
		},
		{
			name:  "Invalid userinfo scope",
			scope: constants.AuthServerResourceIdentifier + ":" + constants.UserinfoPermissionIdentifier,
			expectedError: "The 'authserver:userinfo' scope is automatically included in the access token when an OpenID Connect scope is present. " +
				"There's no need to request it explicitly. Please remove it from your request.",
		},
		{
			name:          "Invalid scope format",
			scope:         "invalid:scope:format",
			expectedError: "Invalid scope format: 'invalid:scope:format'. Scopes must adhere to the resource-identifier:permission-identifier format. For instance: backend-service:create-product.",
		},
		{
			name:  "Valid resource:permission scope",
			scope: "resource1:permission1",
			mockSetup: func() {
				mockDB.On("GetResourceByResourceIdentifier", mock.Anything, "resource1").Return(&models.Resource{Id: 1}, nil)
				mockDB.On("GetPermissionsByResourceId", mock.Anything, int64(1)).Return([]models.Permission{{PermissionIdentifier: "permission1"}}, nil)
			},
			expectedError: "",
		},
		{
			name:  "Invalid resource",
			scope: "invalid-resource:permission",
			mockSetup: func() {
				mockDB.On("GetResourceByResourceIdentifier", mock.Anything, "invalid-resource").Return(nil, nil)
			},
			expectedError: "Invalid scope: 'invalid-resource:permission'. Could not find a resource with identifier 'invalid-resource'.",
		},
		{
			name:  "Invalid permission",
			scope: "resource1:invalid-permission",
			mockSetup: func() {
				mockDB.On("GetResourceByResourceIdentifier", mock.Anything, "resource1").Return(&models.Resource{Id: 1}, nil)
				mockDB.On("GetPermissionsByResourceId", mock.Anything, int64(1)).Return([]models.Permission{{PermissionIdentifier: "valid-permission"}}, nil)
			},
			expectedError: "Scope 'resource1:invalid-permission' is invalid. The resource identified by 'resource1' does not have a permission with identifier 'invalid-permission'.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := validator.ValidateScopes(tt.scope)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				customErr := err.(*customerrors.ErrorDetail)
				assert.Equal(t, tt.expectedError, customErr.GetDescription())
			}
		})
	}
}

func TestValidateScopes_MultipleScopesInSingleRequest(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	mockDB.On("GetResourceByResourceIdentifier", mock.Anything, "resource1").Return(&models.Resource{Id: 1}, nil)
	mockDB.On("GetPermissionsByResourceId", mock.Anything, int64(1)).Return([]models.Permission{{PermissionIdentifier: "permission1"}}, nil)
	mockDB.On("GetResourceByResourceIdentifier", mock.Anything, "resource2").Return(&models.Resource{Id: 2}, nil)
	mockDB.On("GetPermissionsByResourceId", mock.Anything, int64(2)).Return([]models.Permission{{PermissionIdentifier: "permission2"}}, nil)
	scope := "openid profile resource1:permission1 resource2:permission2"
	err := validator.ValidateScopes(scope)
	assert.NoError(t, err)
}

func TestValidateScopes_WithLeadingAndTrailingSpaces(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	mockDB.On("GetResourceByResourceIdentifier", mock.Anything, "resource1").Return(&models.Resource{Id: 1}, nil)
	mockDB.On("GetResourceByResourceIdentifier", mock.Anything, "resource1").Return(&models.Resource{Id: 1}, nil)
	mockDB.On("GetPermissionsByResourceId", mock.Anything, int64(1)).Return([]models.Permission{{PermissionIdentifier: "permission1"}}, nil)
	scope := "  openid  profile  resource1:permission1  "
	err := validator.ValidateScopes(scope)
	assert.NoError(t, err)
}

func TestValidateScopes_MaximumNumberOfScopes(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	// Assuming a theoretical maximum of 100 scopes
	scopes := make([]string, 100)
	for i := 0; i < 100; i++ {
		resourceName := fmt.Sprintf("resource%d", i)
		permissionName := fmt.Sprintf("permission%d", i)
		scopes[i] = fmt.Sprintf("%s:%s", resourceName, permissionName)

		mockDB.On("GetResourceByResourceIdentifier", mock.Anything, resourceName).Return(&models.Resource{Id: int64(i)}, nil)
		mockDB.On("GetPermissionsByResourceId", mock.Anything, int64(i)).Return([]models.Permission{{PermissionIdentifier: permissionName}}, nil)
	}

	scope := strings.Join(scopes, " ")
	err := validator.ValidateScopes(scope)
	assert.NoError(t, err)
}

func TestValidateRequest_InvalidResponseType(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	input := ValidateRequestInput{
		ResponseType:        "token",
		CodeChallengeMethod: "S256",
		CodeChallenge:       "valid_challenge",
	}

	err := validator.ValidateRequest(&input)
	assert.Error(t, err)
	customErr := err.(*customerrors.ErrorDetail)
	assert.Equal(t, "Ensure response_type is set to 'code' as it's the only supported value.", customErr.GetDescription())
}

func TestValidateRequest_InvalidCodeChallengeMethod(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	input := ValidateRequestInput{
		ResponseType:        "code",
		CodeChallengeMethod: "plain",
		CodeChallenge:       "valid_challenge",
	}

	err := validator.ValidateRequest(&input)
	assert.Error(t, err)
	customErr := err.(*customerrors.ErrorDetail)
	assert.Equal(t, "PKCE is required. Ensure code_challenge_method is set to 'S256'.", customErr.GetDescription())
}

func TestValidateRequest_CodeChallengeTooShort(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	input := ValidateRequestInput{
		ResponseType:        "code",
		CodeChallengeMethod: "S256",
		CodeChallenge:       "short",
	}

	err := validator.ValidateRequest(&input)
	assert.Error(t, err)
	customErr := err.(*customerrors.ErrorDetail)
	assert.Equal(t, "The code_challenge parameter is either missing or incorrect. It should be 43 to 128 characters long.", customErr.GetDescription())
}

func TestValidateRequest_CodeChallengeTooLong(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	input := ValidateRequestInput{
		ResponseType:        "code",
		CodeChallengeMethod: "S256",
		CodeChallenge:       string(make([]byte, 129)),
	}

	err := validator.ValidateRequest(&input)
	assert.Error(t, err)
	customErr := err.(*customerrors.ErrorDetail)
	assert.Equal(t, "The code_challenge parameter is either missing or incorrect. It should be 43 to 128 characters long.", customErr.GetDescription())
}

func TestValidateRequest_InvalidResponseMode(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	input := ValidateRequestInput{
		ResponseType:        "code",
		CodeChallengeMethod: "S256",
		CodeChallenge:       "a_valid_code_challenge_that_meets_length_requirements",
		ResponseMode:        "invalid_mode",
	}

	err := validator.ValidateRequest(&input)
	assert.Error(t, err)
	customErr := err.(*customerrors.ErrorDetail)
	assert.Equal(t, "Invalid response_mode parameter. Supported values are: query, fragment, form_post.", customErr.GetDescription())
}

func TestValidateRequest_ValidInput(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	input := ValidateRequestInput{
		ResponseType:        "code",
		CodeChallengeMethod: "S256",
		CodeChallenge:       "a_valid_code_challenge_that_meets_length_requirements",
		ResponseMode:        "query",
	}

	err := validator.ValidateRequest(&input)
	assert.NoError(t, err)
}

func TestValidateRequest_EmptyResponseMode(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewAuthorizeValidator(mockDB)
	input := ValidateRequestInput{
		ResponseType:        "code",
		CodeChallengeMethod: "S256",
		CodeChallenge:       "a_valid_code_challenge_that_meets_length_requirements",
		ResponseMode:        "",
	}

	err := validator.ValidateRequest(&input)
	assert.NoError(t, err)
}
