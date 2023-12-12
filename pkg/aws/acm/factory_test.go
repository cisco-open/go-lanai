package acm

//func TestAwsSessionFactoryImpl_New(t *testing.T) {
//	factory := NewClientFactory(
//		Properties{
//			Region: "us-west-2",
//			Credentials: Credentials{
//				Type:   "static",
//				Id:     "test",
//				Secret: "test",
//			},
//		})
//
//	ctx := context.Background()
//	result, err := factory.New(ctx)
//
//	require.NoError(t, err)
//	assert.IsType(t, &acm.ACM{}, result)
//	bctx := bootstrap.NewApplicationContext()
//	acmclient, e := newDefaultClient(bctx, factory)
//	assert.Nil(t, e)
//	assert.NotNil(t, acmclient)
//
//	factory = NewClientFactory(
//		Properties{
//			Region: "us-east-1",
//			Credentials: Credentials{
//				Type:            "sts",
//				RoleARN:         "some-test-arn",
//				RoleSessionName: "myappname",
//			},
//		})
//
//	result, err = factory.New(ctx)
//
//	require.NoError(t, err)
//	assert.IsType(t, &acm.ACM{}, result)
//	bctx = bootstrap.NewApplicationContext()
//	acmclient, e = newDefaultClient(bctx, factory)
//	assert.Nil(t, e)
//	assert.NotNil(t, acmclient)
//}
