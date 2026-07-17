package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
)

// ErrNotFound is returned when an item does not exist in the bucket.
var ErrNotFound = errors.New("item not found")

// Store is a COS-backed repository for Items.
// Each Item is stored as a JSON object whose key is the item ID.
type Store struct {
	svc    *s3.S3
	bucket string
}

// NewStore creates a Store connected to IBM COS using the supplied config.
func NewStore(cfg COSConfig) (*Store, error) {
	const authEndpoint = "https://iam.cloud.ibm.com/identity/token"

	conf := aws.NewConfig().
		WithEndpoint(cfg.Endpoint).
		WithCredentials(ibmiam.NewStaticCredentials(
			aws.NewConfig(), authEndpoint, cfg.APIKey, cfg.InstanceCRN,
		)).
		WithS3ForcePathStyle(true)

	sess, err := session.NewSession(conf)
	if err != nil {
		return nil, fmt.Errorf("creating COS session: %w", err)
	}

	return &Store{
		svc:    s3.New(sess),
		bucket: cfg.Bucket,
	}, nil
}

// objectKey returns the S3 object key for an item.
func objectKey(id string) string { return "items/" + id + ".json" }

// Create writes a new Item object to COS.
func (s *Store) Create(item *Item) error {
	return s.put(item)
}

// GetByID fetches a single Item from COS by its ID.
func (s *Store) GetByID(id string) (*Item, error) {
	out, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey(id)),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetObject %s: %w", id, err)
	}
	defer out.Body.Close()

	var item Item
	if err := json.NewDecoder(out.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("decoding item %s: %w", id, err)
	}
	return &item, nil
}

// List fetches all Items stored under the "items/" prefix.
func (s *Store) List() ([]*Item, error) {
	out, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String("items/"),
	})
	if err != nil {
		return nil, fmt.Errorf("ListObjectsV2: %w", err)
	}

	items := make([]*Item, 0, len(out.Contents))
	for _, obj := range out.Contents {
		// derive ID from key "items/<id>.json"
		id := (*obj.Key)[len("items/") : len(*obj.Key)-len(".json")]
		item, err := s.GetByID(id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// Update overwrites an existing Item in COS.
func (s *Store) Update(item *Item) error {
	if _, err := s.GetByID(item.ID); err != nil {
		return err
	}
	return s.put(item)
}

// Delete removes the Item object from COS.
func (s *Store) Delete(id string) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	_, err := s.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey(id)),
	})
	if err != nil {
		return fmt.Errorf("DeleteObject %s: %w", id, err)
	}
	return nil
}

// put marshals an Item and writes it to COS via PutObject.
func (s *Store) put(item *Item) error {
	body, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshalling item: %w", err)
	}
	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey(item.ID)),
		Body:        aws.ReadSeekCloser(bytes.NewReader(body)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("PutObject %s: %w", item.ID, err)
	}
	return nil
}

// isNotFound reports whether a COS error is a 404 / NoSuchKey.
func isNotFound(err error) bool {
	type awsErr interface {
		Code() string
	}
	if ae, ok := err.(awsErr); ok {
		return ae.Code() == "NoSuchKey" || ae.Code() == "NotFound"
	}
	return errors.Is(err, io.EOF)
}
