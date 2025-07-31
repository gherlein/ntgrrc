package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ModelDetector contains logic for detecting Netgear switch models
type ModelDetector struct{}

// NewModelDetector creates a new model detector
func NewModelDetector() *ModelDetector {
	return &ModelDetector{}
}

// DetectFromHTML attempts to detect the switch model from HTML content
func (md *ModelDetector) DetectFromHTML(htmlContent string) string {
	// Check for GS316EPP first (more specific than GS316EP)
	if strings.Contains(htmlContent, "GS316EPP") {
		return "GS316EPP"
	}
	
	// Check for GS316EP
	if strings.Contains(htmlContent, "GS316EP") {
		return "GS316EP"
	}
	
	// Check for other specific models
	models := []string{"GS305EP", "GS305EPP", "GS308EP", "GS308EPP"}
	for _, model := range models {
		if strings.Contains(htmlContent, model) {
			return model
		}
	}
	
	// If no specific model found but it looks like a redirect page, assume GS30xEPx
	if strings.Contains(htmlContent, "Redirect to Login") || 
	   strings.Contains(htmlContent, "redirect") {
		return "GS30xEPx"
	}
	
	return ""
}

// POEDataParser contains logic for parsing POE-related data
type POEDataParser struct{}

// NewPOEDataParser creates a new POE data parser
func NewPOEDataParser() *POEDataParser {
	return &POEDataParser{}
}

// ParsePOEStatus parses POE status data from HTML/JavaScript response
func (p *POEDataParser) ParsePOEStatus(content string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	// This is a simplified parser - in real implementation, you'd need to
	// parse the specific JavaScript/HTML format that the switches return
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// Look for POE status tables or JavaScript data
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			if j == 0 {
				return // Skip header row
			}
			
			portData := make(map[string]interface{})
			row.Find("td").Each(func(k int, cell *goquery.Selection) {
				cellText := strings.TrimSpace(cell.Text())
				switch k {
				case 0:
					if portID, err := strconv.Atoi(cellText); err == nil {
						portData["port_id"] = portID
					}
				case 1:
					portData["port_name"] = cellText
				case 2:
					portData["status"] = cellText
				case 3:
					portData["power_class"] = cellText
				case 4:
					if voltage, err := strconv.ParseFloat(cellText, 64); err == nil {
						portData["voltage_v"] = voltage
					}
				case 5:
					if current, err := strconv.ParseFloat(cellText, 64); err == nil {
						portData["current_ma"] = current
					}
				case 6:
					if power, err := strconv.ParseFloat(cellText, 64); err == nil {
						portData["power_w"] = power
					}
				}
			})
			
			if len(portData) > 0 {
				results = append(results, portData)
			}
		})
	})
	
	return results, nil
}

// ParsePOESettings parses POE settings data from HTML/JavaScript response
func (p *POEDataParser) ParsePOESettings(content string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	// Similar to ParsePOEStatus but for settings data
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// Parse POE settings from forms or tables
	doc.Find("form, table").Each(func(i int, element *goquery.Selection) {
		// Extract POE settings based on the specific HTML structure
		// This would need to be implemented based on actual switch responses
		settingsData := make(map[string]interface{})
		
		// Example parsing logic (would need to be adapted for real format)
		element.Find("input, select").Each(func(j int, input *goquery.Selection) {
			name, _ := input.Attr("name")
			value, _ := input.Attr("value")
			
			if name != "" {
				settingsData[name] = value
			}
		})
		
		if len(settingsData) > 0 {
			results = append(results, settingsData)
		}
	})
	
	return results, nil
}

// PortDataParser contains logic for parsing port-related data
type PortDataParser struct{}

// NewPortDataParser creates a new port data parser
func NewPortDataParser() *PortDataParser {
	return &PortDataParser{}
}

// ParsePortSettings parses port settings from HTML content
func (p *PortDataParser) ParsePortSettings(content string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// Parse port settings from tables or forms
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			if j == 0 {
				return // Skip header
			}
			
			portData := make(map[string]interface{})
			row.Find("td").Each(func(k int, cell *goquery.Selection) {
				cellText := strings.TrimSpace(cell.Text())
				switch k {
				case 0:
					if portID, err := strconv.Atoi(cellText); err == nil {
						portData["port_id"] = portID
					}
				case 1:
					portData["port_name"] = cellText
				case 2:
					portData["speed"] = cellText
				case 3:
					portData["ingress_limit"] = cellText
				case 4:
					portData["egress_limit"] = cellText
				case 5:
					portData["flow_control"] = strings.ToLower(cellText) == "on"
				case 6:
					portData["status"] = cellText
				case 7:
					portData["link_speed"] = cellText
				}
			})
			
			if len(portData) > 0 {
				results = append(results, portData)
			}
		})
	})
	
	return results, nil
}

// ExtractSessionToken extracts session token from response content
func ExtractSessionToken(content string) string {
	// Look for SID cookie or session token in various formats
	patterns := []string{
		`SID=([a-fA-F0-9]+)`,
		`sessionid=([a-fA-F0-9]+)`,
		`token["\s]*[:=]["\s]*([a-fA-F0-9]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	
	return ""
}

// ExtractGambitToken extracts Gambit token from response
func ExtractGambitToken(content string) string {
	// Look for Gambit token in JavaScript or HTML
	patterns := []string{
		`Gambit["\s]*[:=]["\s]*([a-fA-F0-9]+)`,
		`gambit["\s]*[:=]["\s]*([a-fA-F0-9]+)`,
		`rand["\s]*[:=]["\s]*([0-9]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	
	return ""
}

// ExtractErrorMessage extracts error messages from response content
func ExtractErrorMessage(content string) string {
	// Look for common error patterns
	patterns := []string{
		`error["\s]*[:=]["\s]*"([^"]+)"`,
		`<div[^>]*error[^>]*>([^<]+)</div>`,
		`alert\s*\(\s*"([^"]+)"\s*\)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	
	return ""
}