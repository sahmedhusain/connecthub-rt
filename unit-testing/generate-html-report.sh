#!/bin/bash

# HTML Test Report Generator
# Generates comprehensive HTML reports from test results
# Version: 1.0.0

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPORT_DIR="$SCRIPT_DIR/test-reports"
COVERAGE_DIR="$SCRIPT_DIR/coverage"
HTML_REPORT_DIR="$SCRIPT_DIR/html-reports"
TEMPLATE_DIR="$SCRIPT_DIR/report-templates"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

log_header() {
    echo -e "${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}"
}

# Help function
show_help() {
    cat << EOF
${CYAN}HTML Test Report Generator${NC}
${CYAN}Version: 1.0.0${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS] [REPORT_TYPE]

${YELLOW}REPORT TYPES:${NC}
    summary             Generate summary report (default)
    detailed            Generate detailed test report
    coverage            Generate coverage report
    all                 Generate all report types

${YELLOW}OPTIONS:${NC}
    -o, --output DIR    Output directory (default: ${HTML_REPORT_DIR})
    -t, --template DIR  Template directory (default: ${TEMPLATE_DIR})
    -v, --verbose       Enable verbose output
    -h, --help          Show this help message

${YELLOW}EXAMPLES:${NC}
    $0                          # Generate summary report
    $0 detailed                 # Generate detailed report
    $0 all --verbose            # Generate all reports with verbose output
    $0 coverage -o ./reports    # Generate coverage report to custom directory

${YELLOW}OUTPUT:${NC}
    HTML reports are saved to: ${HTML_REPORT_DIR}/
    Reports include interactive charts, filtering, and export options

For more information, visit: https://github.com/your-repo/real-time-forum
EOF
}

# Parse command line arguments
parse_arguments() {
    REPORT_TYPE="summary"
    VERBOSE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            summary|detailed|coverage|all)
                REPORT_TYPE="$1"
                shift
                ;;
            -o|--output)
                HTML_REPORT_DIR="$2"
                shift 2
                ;;
            -t|--template)
                TEMPLATE_DIR="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use '$0 --help' for usage information."
                exit 1
                ;;
        esac
    done
}

# Setup environment
setup_environment() {
    log_debug "Setting up environment..."
    
    # Create necessary directories
    mkdir -p "$HTML_REPORT_DIR" "$TEMPLATE_DIR"
    
    # Check if test reports exist
    if [ ! -d "$REPORT_DIR" ]; then
        log_warn "Test reports directory not found: $REPORT_DIR"
        log_info "Run tests first to generate reports"
    fi
    
    log_debug "Environment setup completed"
}

# Create HTML report templates
create_templates() {
    log_info "Creating HTML report templates..."
    
    # Create base template
    cat > "$TEMPLATE_DIR/base.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{TITLE}} - Real-Time Forum Test Report</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; color: #333; }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 2rem; border-radius: 12px; margin-bottom: 2rem; }
        .header h1 { font-size: 2rem; margin-bottom: 0.5rem; }
        .header p { opacity: 0.9; }
        .card { background: white; border-radius: 12px; padding: 1.5rem; margin-bottom: 1.5rem; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .stat-card { background: white; border-radius: 8px; padding: 1rem; text-align: center; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .stat-value { font-size: 2rem; font-weight: bold; margin-bottom: 0.5rem; }
        .stat-label { color: #666; font-size: 0.9rem; }
        .passed { color: #28a745; }
        .failed { color: #dc3545; }
        .skipped { color: #ffc107; }
        .total { color: #17a2b8; }
        .test-results { margin-top: 2rem; }
        .test-item { padding: 1rem; border-left: 4px solid #ddd; margin-bottom: 1rem; background: #f8f9fa; }
        .test-item.passed { border-left-color: #28a745; }
        .test-item.failed { border-left-color: #dc3545; }
        .test-item.skipped { border-left-color: #ffc107; }
        .test-name { font-weight: bold; margin-bottom: 0.5rem; }
        .test-duration { color: #666; font-size: 0.9rem; }
        .footer { text-align: center; margin-top: 3rem; color: #666; }
        @media (max-width: 768px) { .stats-grid { grid-template-columns: 1fr; } }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1><i class="fas fa-chart-bar"></i> {{TITLE}}</h1>
            <p>Generated on {{TIMESTAMP}}</p>
        </div>
        {{CONTENT}}
        <div class="footer">
            <p>Real-Time Forum Test Suite - Generated by HTML Report Generator v1.0.0</p>
        </div>
    </div>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script>{{SCRIPTS}}</script>
</body>
</html>
EOF

    log_info "Templates created successfully"
}

# Generate summary report
generate_summary_report() {
    log_info "Generating summary report..."
    
    local output_file="$HTML_REPORT_DIR/summary.html"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Parse test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    local skipped_tests=0
    
    if [ -d "$REPORT_DIR" ]; then
        # Count results from all test report files
        for report_file in "$REPORT_DIR"/*.txt; do
            if [ -f "$report_file" ]; then
                local file_total=$(grep -c "=== RUN" "$report_file" 2>/dev/null || echo "0")
                local file_passed=$(grep -c "--- PASS:" "$report_file" 2>/dev/null || echo "0")
                local file_failed=$(grep -c "--- FAIL:" "$report_file" 2>/dev/null || echo "0")
                local file_skipped=$(grep -c "--- SKIP:" "$report_file" 2>/dev/null || echo "0")
                
                total_tests=$((total_tests + file_total))
                passed_tests=$((passed_tests + file_passed))
                failed_tests=$((failed_tests + file_failed))
                skipped_tests=$((skipped_tests + file_skipped))
            fi
        done
    fi
    
    # Generate content
    local content="
        <div class=\"stats-grid\">
            <div class=\"stat-card\">
                <div class=\"stat-value total\">$total_tests</div>
                <div class=\"stat-label\">Total Tests</div>
            </div>
            <div class=\"stat-card\">
                <div class=\"stat-value passed\">$passed_tests</div>
                <div class=\"stat-label\">Passed</div>
            </div>
            <div class=\"stat-card\">
                <div class=\"stat-value failed\">$failed_tests</div>
                <div class=\"stat-label\">Failed</div>
            </div>
            <div class=\"stat-card\">
                <div class=\"stat-value skipped\">$skipped_tests</div>
                <div class=\"stat-label\">Skipped</div>
            </div>
        </div>
        
        <div class=\"card\">
            <h2><i class=\"fas fa-chart-pie\"></i> Test Results Overview</h2>
            <canvas id=\"resultsChart\" width=\"400\" height=\"200\"></canvas>
        </div>
        
        <div class=\"card\">
            <h2><i class=\"fas fa-clock\"></i> Recent Test Runs</h2>
            <div class=\"test-results\">
    "
    
    # Add recent test files
    if [ -d "$REPORT_DIR" ]; then
        for report_file in $(ls -t "$REPORT_DIR"/*.txt 2>/dev/null | head -5); do
            if [ -f "$report_file" ]; then
                local filename=$(basename "$report_file")
                local file_passed=$(grep -c "--- PASS:" "$report_file" 2>/dev/null || echo "0")
                local file_failed=$(grep -c "--- FAIL:" "$report_file" 2>/dev/null || echo "0")
                local status_class="passed"
                
                if [ "$file_failed" -gt 0 ]; then
                    status_class="failed"
                fi
                
                content="$content
                <div class=\"test-item $status_class\">
                    <div class=\"test-name\">$filename</div>
                    <div class=\"test-duration\">Passed: $file_passed, Failed: $file_failed</div>
                </div>"
            fi
        done
    fi
    
    content="$content
            </div>
        </div>
    "
    
    # Generate chart script
    local scripts="
        const ctx = document.getElementById('resultsChart').getContext('2d');
        new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Passed', 'Failed', 'Skipped'],
                datasets: [{
                    data: [$passed_tests, $failed_tests, $skipped_tests],
                    backgroundColor: ['#28a745', '#dc3545', '#ffc107']
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: { position: 'bottom' }
                }
            }
        });
    "
    
    # Create final HTML
    sed -e "s/{{TITLE}}/Test Summary Report/g" \
        -e "s/{{TIMESTAMP}}/$timestamp/g" \
        -e "s|{{CONTENT}}|$content|g" \
        -e "s|{{SCRIPTS}}|$scripts|g" \
        "$TEMPLATE_DIR/base.html" > "$output_file"
    
    log_info "Summary report generated: $output_file"
}

# Generate detailed report
generate_detailed_report() {
    log_info "Generating detailed report..."
    
    local output_file="$HTML_REPORT_DIR/detailed.html"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # This would contain detailed test results parsing
    # For now, create a placeholder
    local content="
        <div class=\"card\">
            <h2><i class=\"fas fa-list\"></i> Detailed Test Results</h2>
            <p>Detailed test results will be displayed here.</p>
            <p>This feature will be implemented in the next iteration.</p>
        </div>
    "
    
    # Create final HTML
    sed -e "s/{{TITLE}}/Detailed Test Report/g" \
        -e "s/{{TIMESTAMP}}/$timestamp/g" \
        -e "s|{{CONTENT}}|$content|g" \
        -e "s|{{SCRIPTS}}||g" \
        "$TEMPLATE_DIR/base.html" > "$output_file"
    
    log_info "Detailed report generated: $output_file"
}

# Generate coverage report
generate_coverage_report() {
    log_info "Generating coverage report..."
    
    local output_file="$HTML_REPORT_DIR/coverage.html"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Parse coverage data if available
    local coverage_percentage="0"
    if [ -d "$COVERAGE_DIR" ]; then
        local latest_coverage=$(ls -t "$COVERAGE_DIR"/*_summary.txt 2>/dev/null | head -1)
        if [ -f "$latest_coverage" ]; then
            coverage_percentage=$(tail -1 "$latest_coverage" | awk '{print $3}' | sed 's/%//' || echo "0")
        fi
    fi
    
    local content="
        <div class=\"card\">
            <h2><i class=\"fas fa-chart-area\"></i> Code Coverage</h2>
            <div class=\"stats-grid\">
                <div class=\"stat-card\">
                    <div class=\"stat-value total\">${coverage_percentage}%</div>
                    <div class=\"stat-label\">Overall Coverage</div>
                </div>
            </div>
            <p>Detailed coverage analysis will be available in future versions.</p>
        </div>
    "
    
    # Create final HTML
    sed -e "s/{{TITLE}}/Coverage Report/g" \
        -e "s/{{TIMESTAMP}}/$timestamp/g" \
        -e "s|{{CONTENT}}|$content|g" \
        -e "s|{{SCRIPTS}}||g" \
        "$TEMPLATE_DIR/base.html" > "$output_file"
    
    log_info "Coverage report generated: $output_file"
}

# Generate all reports
generate_all_reports() {
    log_header "Generating All Reports"
    
    generate_summary_report
    generate_detailed_report
    generate_coverage_report
    
    # Create index page
    local index_file="$HTML_REPORT_DIR/index.html"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    local content="
        <div class=\"card\">
            <h2><i class=\"fas fa-home\"></i> Available Reports</h2>
            <div style=\"display: grid; gap: 1rem; margin-top: 1rem;\">
                <a href=\"summary.html\" style=\"display: block; padding: 1rem; background: #f8f9fa; border-radius: 8px; text-decoration: none; color: #333;\">
                    <i class=\"fas fa-chart-bar\"></i> Summary Report
                </a>
                <a href=\"detailed.html\" style=\"display: block; padding: 1rem; background: #f8f9fa; border-radius: 8px; text-decoration: none; color: #333;\">
                    <i class=\"fas fa-list\"></i> Detailed Report
                </a>
                <a href=\"coverage.html\" style=\"display: block; padding: 1rem; background: #f8f9fa; border-radius: 8px; text-decoration: none; color: #333;\">
                    <i class=\"fas fa-chart-area\"></i> Coverage Report
                </a>
            </div>
        </div>
    "
    
    sed -e "s/{{TITLE}}/Test Reports Index/g" \
        -e "s/{{TIMESTAMP}}/$timestamp/g" \
        -e "s|{{CONTENT}}|$content|g" \
        -e "s|{{SCRIPTS}}||g" \
        "$TEMPLATE_DIR/base.html" > "$index_file"
    
    log_info "Index page generated: $index_file"
    log_info "All reports generated successfully!"
}

# Main execution
main() {
    log_header "HTML Test Report Generator"
    
    # Parse arguments
    parse_arguments "$@"
    
    # Setup environment
    setup_environment
    
    # Create templates
    create_templates
    
    # Generate reports based on type
    case "$REPORT_TYPE" in
        "summary")
            generate_summary_report
            ;;
        "detailed")
            generate_detailed_report
            ;;
        "coverage")
            generate_coverage_report
            ;;
        "all")
            generate_all_reports
            ;;
        *)
            log_error "Unknown report type: $REPORT_TYPE"
            show_help
            exit 1
            ;;
    esac
    
    log_info "Report generation completed!"
    log_info "Open $HTML_REPORT_DIR/index.html in your browser to view reports"
}

# Run main function
main "$@"
