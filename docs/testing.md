# Testing Documentation

This document outlines the test cases implemented in the Vongga Platform backend.

## Post Use Case Tests

### CreatePost Tests
1. **Success Case**
   - Creates a new post with all fields (content, media, tags, location, visibility)
   - Verifies all fields are correctly set
   - Checks initial values (comment count, subpost count, etc.)
   - Validates post is not marked as edited

2. **Repository Error Case**
   - Tests error handling when repository fails
   - Ensures error is propagated correctly
   - Verifies no post is created on error

### UpdatePost Tests
1. **Success Case**
   - Updates all post fields (content, media, tags, location, visibility)
   - Verifies edit history is correctly maintained
   - Checks post is marked as edited
   - Validates all fields are updated correctly

2. **Post Not Found Case**
   - Tests behavior when updating non-existent post
   - Verifies appropriate error is returned

3. **Update Error Case**
   - Tests error handling during update operation
   - Ensures error is propagated correctly

### DeletePost Tests
1. **Success Case**
   - Deletes post with no subposts
   - Verifies deletion is successful

2. **Post Not Found Case**
   - Tests deletion of non-existent post
   - Verifies appropriate error is returned

3. **Has SubPosts Case**
   - Tests deletion of post with subposts
   - Verifies deletion is prevented
   - Ensures appropriate error is returned

### GetPost Tests
1. **Success Without SubPosts Case**
   - Retrieves post without subposts
   - Verifies all post fields are correct

2. **Success With SubPosts Case**
   - Retrieves post with subposts
   - Verifies both post and subpost data
   - Validates subpost ordering

3. **Post Not Found Case**
   - Tests retrieval of non-existent post
   - Verifies appropriate error is returned

4. **SubPosts Error Case**
   - Tests error handling during subpost retrieval
   - Ensures error is propagated correctly

### ListPosts Tests
1. **Success Without SubPosts Case**
   - Lists posts for a user without subposts
   - Verifies pagination (limit, offset)
   - Validates post data

2. **Success With SubPosts Case**
   - Lists posts with their subposts
   - Verifies both post and subpost data
   - Validates subpost ordering for each post

3. **Repository Error Case**
   - Tests error handling from repository
   - Ensures error is propagated correctly

4. **SubPosts Error Case**
   - Tests error handling during subpost retrieval
   - Ensures error is propagated correctly

## Test Coverage

All test cases cover:
- Success scenarios
- Error handling
- Edge cases
- Data validation
- Repository interactions
- Business logic validation

## Mock Implementations

### PostRepository Mock
- Create
- Update
- Delete
- FindByID
- FindByUserID
- FindSubPosts

### SubPostRepository Mock
- Create
- Update
- Delete
- FindByID
- FindByPostID
- FindByParentID
- UpdateOrder

## Running Tests

To run all tests:
```bash
go test ./usecase/... -v
```

To run specific test:
```bash
go test ./usecase/... -v -run TestPostUseCase_CreatePost
```
