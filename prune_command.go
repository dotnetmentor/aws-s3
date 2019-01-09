package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/cli"
	"gopkg.in/cheggaaa/pb.v1"
)

func pruneCommandFactory() (cli.Command, error) {
	c := &pruneCommand{}
	c.Flags = flag.NewFlagSet("prune", flag.ExitOnError)
	c.Bucket = c.Flags.String("bucket", "", "Bucket name")
	c.Region = c.Flags.String("region", "eu-central-1", "Region")
	c.Prefix = c.Flags.String("prefix", "", "Key prefix")
	c.MaxAge = c.Flags.Duration("max-age", 24*time.Hour, "Max age")
	c.MaxFiles = c.Flags.Int("max-files", 10000, "Max files to process")
	c.DryRun = c.Flags.Bool("dry-run", false, "Dry run")
	c.ShowProgress = c.Flags.Bool("progress", false, "Show progress")
	c.ListFiles = c.Flags.Bool("list-files", false, "List files")
	c.ListFolders = c.Flags.Bool("list-folders", true, "List folders")
	return c, nil
}

// PruneCommand removes old s3 objects
type pruneCommand struct {
	Flags        *flag.FlagSet
	Bucket       *string
	Region       *string
	Prefix       *string
	MaxAge       *time.Duration
	MaxFiles     *int
	DryRun       *bool
	ShowProgress *bool
	ListFiles    *bool
	ListFolders  *bool
}

func (c *pruneCommand) Run(args []string) int {
	c.Flags.Parse(args)

	bucket := *c.Bucket
	region := *c.Region
	prefix := *c.Prefix
	maxAge := *c.MaxAge
	maxFiles := *c.MaxFiles
	dryrun := *c.DryRun
	showProgress := *c.ShowProgress

	if bucket == "" || region == "" || prefix == "" {
		c.Flags.PrintDefaults()
		os.Exit(1)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	svc := s3.New(sess)

	start := time.Now()
	fmt.Printf("searching s3 (bucket=%s region=%s prefix=%s max-age=%s max-files=%v dry-run=%v)\n", bucket, region, prefix, maxAge, maxFiles, dryrun)

	objs, err := searchObjects(svc, bucket, region, prefix, maxFiles, showProgress)
	if err == nil {
		sortObjects(objs)
		removeObjs, keepObjs := processObjects(objs, start, maxAge)
		removeObjects(svc, bucket, removeObjs, dryrun, *c.ListFiles, *c.ListFolders)
		keepObjects(keepObjs, dryrun, *c.ListFiles, *c.ListFolders)
		fmt.Printf("completed in %v seconds", time.Since(start).Seconds())
	} else {
		fmt.Println(err)
	}

	return 0
}

func (c *pruneCommand) Help() string {
	fmt.Println("Available flags are:")
	c.Flags.PrintDefaults()
	return ""
}

func (c *pruneCommand) Synopsis() string {
	return "Runs cleanup of s3 objects"
}

func removeObjects(svc *s3.S3, bucket string, objs []Object, dryrun bool, listFiles bool, listDirectories bool) {
	batchSize := 1000
	folders := countFolders(objs)

	if dryrun {
		fmt.Printf("would remove %v s3 objects from %v folders\n", len(objs), folders)
		printObjects(objs, listFiles, listDirectories)
		return
	}

	fmt.Printf("removing %v s3 objects from %v folders in batches of %v\n", len(objs), folders, batchSize)
	printObjects(objs, listFiles, listDirectories)

	if len(objs) == 0 {
		return
	}

	ids := make([]*s3.ObjectIdentifier, 0)
	for _, obj := range objs {
		ids = append(ids, &s3.ObjectIdentifier{
			Key: aws.String(obj.Key),
		})
	}

	batches := make([][]*s3.ObjectIdentifier, 0)
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize

		if end > len(ids) {
			end = len(ids)
		}

		batches = append(batches, ids[i:end])
	}

	for _, batch := range batches {
		removeObjectsBatch(svc, bucket, batch)
	}
}

func removeObjectsBatch(svc *s3.S3, bucket string, ids []*s3.ObjectIdentifier) {
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: ids,
			Quiet:   aws.Bool(false),
		},
	}

	result, err := svc.DeleteObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Printf("successfully deleted %v s3 objects, %v error(s)\n", len(result.Deleted), len(result.Errors))
}

func keepObjects(objs []Object, dryrun bool, listFiles bool, listDirectories bool) {
	folders := countFolders(objs)

	if dryrun {
		fmt.Printf("would keep %v s3 objects in %v folders\n", len(objs), folders)
	} else {
		fmt.Printf("keeping %v s3 objects in %v folders\n", len(objs), folders)
	}

	printObjects(objs, listFiles, listDirectories)
}

func searchObjects(svc *s3.S3, bucket string, region string, prefix string, maxFiles int, showProgress bool) ([]Object, error) {
	var bar *pb.ProgressBar

	allObjs := make([]Object, 0)

	i := 0
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		i++

		if showProgress {
			size := len(p.Contents)
			if (len(allObjs) + len(p.Contents)) > maxFiles {
				size = maxFiles - len(allObjs)
			}
			bar = pb.StartNew(size).Prefix(fmt.Sprintf("  page %d:", i))
		}

		for _, obj := range p.Contents {
			objData, err := svc.HeadObject(&s3.HeadObjectInput{
				Bucket: &bucket,
				Key:    obj.Key,
			})

			if err != nil {
				return false
			}

			allObjs = append(allObjs, Object{
				Bucket:       bucket,
				Key:          *obj.Key,
				LastModified: *objData.LastModified,
			})

			if showProgress {
				bar.Increment()
			}

			if len(allObjs) >= maxFiles {
				return false
			}
		}

		if showProgress {
			bar.Finish()
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	return allObjs, nil
}

func processObjects(objs []Object, start time.Time, maxAge time.Duration) ([]Object, []Object) {
	removeObjs := make([]Object, 0)
	keepObjs := make([]Object, 0)

	for _, obj := range objs {
		parts := strings.Split(obj.Key, "/")
		obj.File = parts[len(parts)-1]
		obj.Folder = strings.Join(parts[:len(parts)-1], "/")
		obj.Age = start.Sub(obj.LastModified)

		if obj.Age > maxAge {
			removeObjs = append(removeObjs, obj)
		} else {
			keepObjs = append(keepObjs, obj)
		}
	}

	return removeObjs, keepObjs
}
