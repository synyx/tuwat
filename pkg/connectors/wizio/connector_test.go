package wizio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

func TestConnector(t *testing.T) {
	alerts := testCollection(t, mockResponse)

	if alerts == nil || len(alerts) != 2 {
		t.Error("There should be 2 alerts")
	}
}

func TestThreat(t *testing.T) {
	alerts := testCollection(t, mockThreatResponse)

	if alerts == nil || len(alerts) != 1 {
		t.Error("There should be 1 alert")
	}
}

func testCollection(t *testing.T, response string) []connectors.Alert {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte(response))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer func() { testServer.Close() }()

	cfg := Config{
		Tag: "test",
		HTTPConfig: common.HTTPConfig{
			URL: testServer.URL,
		},
	}

	var connector connectors.Connector = NewConnector(&cfg)
	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	return alerts
}

func TestDecode(t *testing.T) {
	var foo issuesResponse
	err := json.Unmarshal([]byte(mockResponse), &foo)
	if err != nil {
		t.Fatal(err)
	}
}

const mockResponse = `
{
  "data": {
    "issuesV2": {
      "nodes": [
        {
          "id": "0b33b35a-55e4-45c7-a918-98ec2c271b4c",
          "control": null,
          "createdAt": "2025-01-01T14:06:30.100429Z",
          "updatedAt": "2025-01-03T14:21:45.596548Z",
          "dueAt": null,
          "project": null,
          "status": "OPEN",
          "severity": "HIGH",
          "entity": {
            "id": "370184b8-0994-4403-8d80-46474ef8e294",
            "name": "statefulset-name",
            "type": "STATEFUL_SET"
          },
          "entitySnapshot": {
            "id": "370184b8-0994-4403-8d80-46474ef8e294",
            "type": "STATEFUL_SET",
            "name": "statefulset-name",
            "status": null,
            "cloudPlatform": "Kubernetes",
            "region": ""
          },
          "note": "",
          "serviceTickets": null,
          "sourceRules": [
            {
              "__typename": "CloudEventRule",
              "id": "cer-correlation-id-291",
              "name": "Anomalous scanning tool was executed",
              "description": "Anomalous network manipulation/scanning tool was executed. This could indicate the presence of a network scanner attempting to identify or spread to other resources.",
              "sourceType": "WIZ_SENSOR",
              "type": "CORRELATION",
              "risks": null,
              "securitySubCategories": []
            }
          ]
        },
        {
          "id": "cc056762-8287-4a80-85ee-073b42117e6c",
          "control": {
            "id": "wc-id-1397",
            "name": "Container using an image with initial access vulnerabilities that were validated in runtime"
          },
          "createdAt": "2025-04-01T22:16:53.808853Z",
          "updatedAt": "2025-04-03T10:43:42.659447Z",
          "dueAt": null,
          "project": null,
          "status": "OPEN",
          "severity": "LOW",
          "entity": {
            "id": "1b00ce0b-bd3a-431e-88e0-d0bec847d436",
            "name": "deployment-name",
            "type": "DEPLOYMENT"
          },
          "entitySnapshot": {
            "id": "1b00ce0b-bd3a-431e-88e0-d0bec847d436",
            "type": "DEPLOYMENT",
            "name": "deployment-name",
            "status": "Active",
            "cloudPlatform": "Kubernetes",
            "region": ""
          },
          "note": "",
          "serviceTickets": null,
          "sourceRules": [
            {
              "__typename": "Control",
              "id": "wc-id-1397",
              "name": "Container using an image with initial access vulnerabilities that were validated in runtime",
              "description": "This container has a critical/high severity [initial access vulnerability](https://docs.wiz.io/wiz-docs/docs/vulnerability-findings?lng=en#initial-access). The vulnerable technologies were detected running by the Wiz Runtime Sensor.\\n\\nan attacker might be able to execute code on the publicly exposed resource by exploiting the vulnerability, as long as the conditions for exploitation are met.",
              "resolutionRecommendation": "### Patch vulnerabilities\\n* Update all software running in your environment to the latest version.\\n   * For public images, ensure that you update to the latest version.\\n   * For private images, check the Finding for detailed information regarding the vulnerability.\\n   * Build a new image with the fixed vulnerability and replace it in the resource.\\n* If you cannot use the latest version, prioritize patching resources in your environment according to the attack surface they are exposed to and the potential impact of this resource’s compromise (based on the Issue severity).",
              "risks": [
                "INSECURE_KUBERNETES_CLUSTER",
                "VULNERABILITY"
              ],
              "securitySubCategories": [
                {
                  "title": "Vulnerable resource",
                  "category": {
                    "name": "Vulnerability Assessment",
                    "framework": {
                      "name": "Wiz for Risk Assessment"
                    }
                  }
                },
                {
                  "title": "OPS-20 Managing Vulnerabilities, Malfunctions and Errors - Measurements, Analyses and Assessments of Procedures",
                  "category": {
                    "name": "5.6 Operations (OPS)",
                    "framework": {
                      "name": "C5 - Cloud Computing Compliance Criteria Catalogue"
                    }
                  }
                },
                {
                  "title": "7.1 Establish and Maintain a Vulnerability Management Process",
                  "category": {
                    "name": "7 Continuous Vulnerability Management",
                    "framework": {
                      "name": "CIS Controls v8"
                    }
                  }
                },
                {
                  "title": "IAC-20.4 Dedicated Administrative Machines",
                  "category": {
                    "name": "16-IAC Identification & Authentication",
                    "framework": {
                      "name": "SCF (Secure Controls Framework)"
                    }
                  }
                },
                {
                  "title": "3.11.3 Remediate vulnerabilities in accordance with risk assessments",
                  "category": {
                    "name": "3.11 Risk Assessment",
                    "framework": {
                      "name": "NIST 800-171 Rev.2"
                    }
                  }
                },
                {
                  "title": "Container Security",
                  "category": {
                    "name": "Container Security",
                    "framework": {
                      "name": "Wiz (Legacy)"
                    }
                  }
                },
                {
                  "title": "16.13 Conduct Application Penetration Testing",
                  "category": {
                    "name": "16 Application Software Security",
                    "framework": {
                      "name": "CIS Controls v8"
                    }
                  }
                },
                {
                  "title": "OPS-22 Testing and Documentation of known Vulnerabilities",
                  "category": {
                    "name": "5.6 Operations (OPS)",
                    "framework": {
                      "name": "C5 - Cloud Computing Compliance Criteria Catalogue"
                    }
                  }
                },
                {
                  "title": "7.3 Perform Automated Operating System Patch Management",
                  "category": {
                    "name": "7 Continuous Vulnerability Management",
                    "framework": {
                      "name": "CIS Controls v8"
                    }
                  }
                },
                {
                  "title": "PSS-11 Images for Virtual Machines and Containers",
                  "category": {
                    "name": "5.17 Product Safety and Security (PSS)",
                    "framework": {
                      "name": "C5 - Cloud Computing Compliance Criteria Catalogue"
                    }
                  }
                },
                {
                  "title": "B4.d System security - Vulnerability Management",
                  "category": {
                    "name": "B4 System Security",
                    "framework": {
                      "name": "CAF (Cyber Assessment Framework by NCSC)"
                    }
                  }
                },
                {
                  "title": "Vulnerable Container",
                  "category": {
                    "name": "Data Plane and Workload Security",
                    "framework": {
                      "name": "Wiz for Container & Kubernetes Security"
                    }
                  }
                },
                {
                  "title": "ID.RA-01 Vulnerabilities in assets are identified, validated, and recorded",
                  "category": {
                    "name": "ID.RA Risk Assessment",
                    "framework": {
                      "name": "NIST CSF v2.0"
                    }
                  }
                },
                {
                  "title": "500.05(b) Bi-annual vulnerability assessments, including any systematic scans or reviews of Information Systems reasonably designed to identify publicly known cybersecurity vulnerabilities in the Covered Entity’s Information Systems based on the Risk Assessment.",
                  "category": {
                    "name": "500.05 Penetration Testing and Vulnerability Assessments",
                    "framework": {
                      "name": "NYDFS (23 NYCRR 500)"
                    }
                  }
                },
                {
                  "title": "OPS-19 Managing Vulnerabilities, Malfunctions and Errors - Penetration Tests",
                  "category": {
                    "name": "5.6 Operations (OPS)",
                    "framework": {
                      "name": "C5 - Cloud Computing Compliance Criteria Catalogue"
                    }
                  }
                },
                {
                  "title": "Vulnerable container",
                  "category": {
                    "name": "Container & Kubernetes Security",
                    "framework": {
                      "name": "Wiz for Risk Assessment"
                    }
                  }
                },
                {
                  "title": "T1203 Exploitation for Client Execution",
                  "category": {
                    "name": "TA0002 Execution",
                    "framework": {
                      "name": "MITRE ATT&CK Matrix"
                    }
                  }
                },
                {
                  "title": "Vulnerability Assessment",
                  "category": {
                    "name": "Vulnerability Assessment",
                    "framework": {
                      "name": "Wiz (Legacy)"
                    }
                  }
                },
                {
                  "title": "500.05(a) Annual Penetration Testing of the Covered Entity’s Information Systems determined each given year based on relevant identified risks in accordance with the Risk Assessment",
                  "category": {
                    "name": "500.05 Penetration Testing and Vulnerability Assessments",
                    "framework": {
                      "name": "NYDFS (23 NYCRR 500)"
                    }
                  }
                },
                {
                  "title": "COS-01 Technical safeguards",
                  "category": {
                    "name": "5.9 Communication Security (COS)",
                    "framework": {
                      "name": "C5 - Cloud Computing Compliance Criteria Catalogue"
                    }
                  }
                },
                {
                  "title": "16.3 Perform Root Cause Analysis on Security Vulnerabilities",
                  "category": {
                    "name": "16 Application Software Security",
                    "framework": {
                      "name": "CIS Controls v8"
                    }
                  }
                },
                {
                  "title": "9.4 As part of the ICT risk management framework referred to in Article 6(1), financial entities shall:\\n\\n(a) develop and document an information security policy defining rules to protect the availability, authenticity, integrity and confidentiality of data, information assets and ICT assets, including those of their customers, where applicable;\\n\\n(b) following a risk-based approach, establish a sound network and infrastructure management structure using appropriate techniques, methods and protocols that may include implementing automated mechanisms to isolate affected information assets in the event of cyber-attacks;\\n\\n(c) implement policies that limit the physical or logical access to information assets and ICT assets to what is required for legitimate and approved functions and activities only, and establish to that end a set of policies, procedures and controls that address access rights and ensure a sound administration thereof;\\n\\n(d) implement policies and protocols for strong authentication mechanisms, based on relevant standards and dedicated control systems, and protection measures of cryptographic keys whereby data is encrypted based on results of approved data classification and ICT risk assessment processes;\\n\\n(e) implement documented policies, procedures and controls for ICT change management, including changes to software, hardware, firmware components, systems or security parameters, that are based on a risk assessment approach and are an integral part of the financial entity’s overall change management process, in order to ensure that all changes to ICT systems are recorded, tested, assessed, approved, implemented and verified in a controlled manner;\\n\\n(f) have appropriate and comprehensive documented policies for patches and updates.\\n\\nFor the purposes of the first subparagraph, point (b), financial entities shall design the network connection infrastructure in a way that allows it to be instantaneously severed or segmented in order to minimise and prevent contagion, especially for interconnected financial processes.\\n\\nFor the purposes of the first subparagraph, point (e), the ICT change management process shall be approved by appropriate lines of management and shall have specific protocols in place.",
                  "category": {
                    "name": "Art 9 Protection and prevention - CHAPTER II - ICT risk management",
                    "framework": {
                      "name": "Digital Operational Resilience Act (DORA)"
                    }
                  }
                },
                {
                  "title": "10.m Control of Technical Vulnerabilities",
                  "category": {
                    "name": "10.06 Technical Vulnerability Management - Information Systems Acquisition, Development, and Maintenance",
                    "framework": {
                      "name": "HITRUST CSF v11.2"
                    }
                  }
                },
                {
                  "title": "VPM-06.9 Correlate Scanning Information",
                  "category": {
                    "name": "32-VPM  Vulnerability & Patch Management",
                    "framework": {
                      "name": "SCF (Secure Controls Framework)"
                    }
                  }
                },
                {
                  "title": "21.2.1 The measures to protect network and information systems shall include policies on risk analysis and information system security",
                  "category": {
                    "name": "Article 21 Cybersecurity risk-management measures",
                    "framework": {
                      "name": "NIS2 Directive (Article 21)"
                    }
                  }
                },
                {
                  "title": "SEC06-BP01 Perform vulnerability management",
                  "category": {
                    "name": "SEC 6 Infrastructure protection - How do you protect your compute resources?",
                    "framework": {
                      "name": "AWS Well-Architected Framework (Section 2 - Security)"
                    }
                  }
                },
                {
                  "title": "Validated in runtime vulnerabilities",
                  "category": {
                    "name": "Vulnerability Assessment",
                    "framework": {
                      "name": "Wiz for Risk Assessment"
                    }
                  }
                }
              ]
            }
          ]
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "endCursor": "cmVhbGx5bG9uZ2N1cnNvcnN0cmluZw=="
      }
    }
  }
}
`

const mockThreatResponse = `
{
  "data": {
    "issuesV2": {
      "nodes": [
        {
          "id": "dc14fcff-549d-5d3c-906d-f505ac5b5964",
          "createdAt": "2025-09-05T14:54:42.829455Z",
          "updatedAt": "2025-09-05T14:58:28.023204Z",
          "projects": [
            {
              "id": "17c260a0-aa6d-5246-86c8-e24b700b3ca1",
              "name": "SysOps"
            },
            {
              "id": "bde2f9cf-9275-5e6b-9fa0-e274c57c5641",
              "name": "Contargo"
            },
            {
              "id": "caf67dbd-2f43-5144-a086-ae62689734b2",
              "name": "Stage"
            }
          ],
          "status": "OPEN",
          "severity": "HIGH",
          "entitySnapshot": {
            "id": "b0c8f29e-5a4a-5182-b53a-c003a527c23c",
            "type": "VIRTUAL_MACHINE",
            "name": "vm-3",
            "status": null,
            "kubernetesClusterName": "customer-stage",
            "kubernetesNamespaceName": "",
            "tags": {}
          },
          "notes": [
            {
              "text": "Threat Bewertet",
              "createdAt": "2025-09-10T12:08:26.942477Z",
              "user": {
                "id": "entraid_user",
                "name": "Buch, Jonathan"
              }
            }
          ],
          "serviceTickets": null,
          "sourceRules": [
            {
              "__typename": "CloudEventRule",
              "id": "cer-correlation-id-287",
              "name": "Multiple runtime detections on a resource in a short period of time",
              "description": "Multiple runtime detections were triggered on the same resource in a short period of time. This can indicate the execution of a malicious tool or a script on the resource."
            }
          ]
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "endCursor": "ey"
      }
    }
  }
}`
