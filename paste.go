package paste

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
)

type Replacer func(string) (string, error)

func NewReplacer(s client.ConfigProvider) Replacer {
	svc := ssm.New(s)
	return func(name string) (string, error) {
		out, err := svc.GetParameter(&ssm.GetParameterInput{
			Name:           aws.String(name),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return "", errors.Wrapf(err, "can't retrieve parameter '%v'", name)
		}

		return *out.Parameter.Value, nil
	}
}

func FakeReplacer(_ string) (string, error) { return "---", nil }

const RgxStr = `.*(<%.*%>).*`

func ReplaceAll(rpl Replacer, in io.Reader, out io.Writer) error {
	rgx, err := regexp.Compile(RgxStr)
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(in)

	for sc.Scan() {
		line := sc.Text()

		match, exists := GetPlaceholder(rgx, line)
		if exists {
			value, err := rpl(match)
			if err != nil {
				return err
			}

			line = strings.ReplaceAll(line, "<%"+match+"%>", value)
		}

		_, err = out.Write([]byte(line + "\n"))
		if err != nil {
			return errors.Wrap(err, "can't write line '%v' to buffer")
		}
	}

	return nil
}

func GetPlaceholder(rgx *regexp.Regexp, line string) (string, bool) {
	res := rgx.FindStringSubmatch(line)

	if len(res) < 2 {
		return "", false
	}

	match := res[1]
	match = strings.TrimPrefix(match, "<%")
	match = strings.TrimSuffix(match, "%>")

	return match, true
}
