package reporter

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

type junitTestSuites struct {
	XMLName xml.Name     `xml:"testsuites"`
	Suites  []junitSuite `xml:"testsuite"`
}

type junitSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Cases    []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	File      string        `xml:"file,attr,omitempty"`
	Line      string        `xml:"line,attr,omitempty"`
	Failure   *junitFailure `xml:"failure,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

func WriteJUnit(path string, report *finding.Report) error {
	if report == nil {
		return fmt.Errorf("report is required")
	}
	cases := make([]junitTestCase, 0, len(report.Findings))
	for _, item := range report.Findings {
		body := item.Description
		if item.Recommendation != "" {
			body += "\nRecommendation: " + item.Recommendation
		}
		cases = append(cases, junitTestCase{
			Name: item.RuleID, Classname: string(item.Domain) + "." + item.Category,
			File: item.Location.File, Line: strconv.Itoa(item.Location.Line),
			Failure: &junitFailure{Message: item.Description, Type: string(item.Severity), Body: body},
		})
	}
	document := junitTestSuites{Suites: []junitSuite{{
		Name: report.Project, Tests: len(cases), Failures: len(cases), Cases: cases,
	}}}
	data, err := xml.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JUnit report: %w", err)
	}
	data = append([]byte(xml.Header), append(data, '\n')...)
	return writeAtomic(path, data, "JUnit report")
}
