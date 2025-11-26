// ---------------------------------------------------------
// HELPER: Physical Rotation Logic
// ---------------------------------------------------------

// Rotate X: Top Face -> Front Face
// (Rotates 90deg around the X-axis)
module spin_x (input wire [11:0] i, output wire [11:0] o);
    assign o = {
        i[3],  
        i[7],  
        i[11],  
        i[4],  
        
        i[2],  
        i[10], 
        i[8],  
        i[0],  
        
        i[1],  
        i[6], 
        i[9],  
        i[5]  
    };
endmodule

// Rotate Z: Front Face -> Right Face
// (Rotates 90deg Clockwise around the Z-axis)
module spin_z (input wire [11:0] i, output wire [11:0] o);
    assign o = {
        i[10], // 11 <- 10
        i[9],  // 10 <- 9
        i[8],  // 9  <- 8
        i[11], // 8  <- 11 (Top Ring)

        i[6],  // 7 <- 6
        i[5],  // 6 <- 5
        i[4],  // 5 <- 4
        i[7],  // 4 <- 7   (Middle Ring)

        i[2],  // 3 <- 2
        i[1],  // 2 <- 1
        i[0],  // 1 <- 0
        i[3]   // 0 <- 3   (Bottom Ring)
    };
endmodule

// Rotate Z Counter-Clockwise
module spin_z_cc (input wire [11:0] i, output wire [11:0] o);
    assign o = {
        i[8],  // 11 <- 8
        i[11], // 10 <- 11
        i[10], // 9  <- 10
        i[9],  // 8  <- 9
        
        i[4],  // 7 <- 4
        i[7],  // 6 <- 7
        i[6],  // 5 <- 6
        i[5],  // 4 <- 5
        
        i[0],  // 3 <- 0
        i[3],  // 2 <- 3
        i[2],  // 1 <- 2
        i[1]   // 0 <- 1
    };
endmodule


module is_canonical (
    input  wire [11:0] cube_in,
    output wire        is_canon
);
    wire [11:0] r0, r1, r2, r3, r4, r5;
    
    // --- Generate 6 Base Faces ---
    
    // 1. Top (Original)
    assign r0 = cube_in;

    // 2. Back (Rotate X)
    spin_x x_to_back (.i(r0), .o(r1));

    // 3. Bottom (Rotate X again)
    spin_x x_to_bot  (.i(r1), .o(r2));

    // 4. Front (Rotate X again)
    spin_x x_to_front(.i(r2), .o(r3));

    // 5. Left (Rotate Z then X)
    // spin_z moves Left->Front. spin_x moves Front->Top.
    // So this puts the LEFT face on Top.
    wire [11:0] temp_left;
    spin_z z_pre_left (.i(r0),        .o(temp_left));
    spin_x x_to_left  (.i(temp_left), .o(r4));

    // 6. Right (Rotate Z_CC then X)
    // spin_z_cc moves Right->Front. spin_x moves Front->Top.
    // So this puts the RIGHT face on Top.
    wire [11:0] temp_right;
    spin_z_cc z_pre_right (.i(r0),         .o(temp_right));
    spin_x    x_to_right  (.i(temp_right), .o(r5));


    // --- Generate 3 Z-Rotations for ALL 6 Faces ---
    wire [11:0] r0_90, r0_180, r0_270;
    spin_z z0_1 (.i(r0),     .o(r0_90));
    spin_z z0_2 (.i(r0_90),  .o(r0_180));
    spin_z z0_3 (.i(r0_180), .o(r0_270));

    wire [11:0] r1_90, r1_180, r1_270;
    spin_z z1_1 (.i(r1),     .o(r1_90));
    spin_z z1_2 (.i(r1_90),  .o(r1_180));
    spin_z z1_3 (.i(r1_180), .o(r1_270));

    wire [11:0] r2_90, r2_180, r2_270;
    spin_z z2_1 (.i(r2),     .o(r2_90));
    spin_z z2_2 (.i(r2_90),  .o(r2_180));
    spin_z z2_3 (.i(r2_180), .o(r2_270));

    wire [11:0] r3_90, r3_180, r3_270;
    spin_z z3_1 (.i(r3),     .o(r3_90));
    spin_z z3_2 (.i(r3_90),  .o(r3_180));
    spin_z z3_3 (.i(r3_180), .o(r3_270));

    wire [11:0] r4_90, r4_180, r4_270;
    spin_z z4_1 (.i(r4),     .o(r4_90));
    spin_z z4_2 (.i(r4_90),  .o(r4_180));
    spin_z z4_3 (.i(r4_180), .o(r4_270));

    wire [11:0] r5_90, r5_180, r5_270;
    spin_z z5_1 (.i(r5),     .o(r5_90));
    spin_z z5_2 (.i(r5_90),  .o(r5_180));
    spin_z z5_3 (.i(r5_180), .o(r5_270));

    // --- The Comparator ---
    assign is_canon = 
        (cube_in <= r0_90) && (cube_in <= r0_180) && (cube_in <= r0_270) &&
        (cube_in <= r1) && (cube_in <= r1_90) && (cube_in <= r1_180) && (cube_in <= r1_270) &&
        (cube_in <= r2) && (cube_in <= r2_90) && (cube_in <= r2_180) && (cube_in <= r2_270) &&
        (cube_in <= r3) && (cube_in <= r3_90) && (cube_in <= r3_180) && (cube_in <= r3_270) &&
        (cube_in <= r4) && (cube_in <= r4_90) && (cube_in <= r4_180) && (cube_in <= r4_270) &&
        (cube_in <= r5) && (cube_in <= r5_90) && (cube_in <= r5_180) && (cube_in <= r5_270);
endmodule // is_canonical


// ---------------------------------------------------------
// MODULE 1: Connectivity Checker (The Flood Fill Engine)
// ---------------------------------------------------------
module cube_connectivity (
    input  wire        clk,
    input  wire        rst,
    input  wire        start,
    input  wire [11:0] edges_in,  // The current cube to test
    output reg         is_connected,
    output reg         done,
    output reg         busy
);

    // --- Neighbor Lookup Table (Derived from your Go 'nmask') ---
    // Maps an edge index (0-11) to a bitmask of its neighbors
    function [11:0] get_neighbors(input [3:0] edge_idx);
        case (edge_idx)
            4'd0:  get_neighbors = 12'h03a;
            4'd1:  get_neighbors = 12'h065;
            4'd2:  get_neighbors = 12'h0ca;
            4'd3:  get_neighbors = 12'h095;
            4'd4:  get_neighbors = 12'h909;
            4'd5:  get_neighbors = 12'h303;
            4'd6:  get_neighbors = 12'h606;
            4'd7:  get_neighbors = 12'hc0c;
            4'd8:  get_neighbors = 12'ha30;
            4'd9:  get_neighbors = 12'h560;
            4'd10: get_neighbors = 12'hac0;
            4'd11: get_neighbors = 12'h590;
            default: get_neighbors = 12'h000;
        endcase
    endfunction

    // --- State Machine States ---
    localparam S_IDLE   = 2'd0;
    localparam S_SEED   = 2'd1;
    localparam S_FLOOD  = 2'd2;
    localparam S_CHECK  = 2'd3;

    reg [1:0] state;
    reg [11:0] visited;
    reg [11:0] next_visited;
    
    // --- Optimized Parallel "Flood" Logic ---
    // This generates the "Next Visited" mask efficiently using 12 parallel lookups
    wire [11:0] neighbor_matrix [11:0];
    wire [11:0] active_neighbors [11:0];

    genvar g;
    generate
        for (g = 0; g < 12; g = g + 1) begin : gen_map
            // 1. Get neighbors for this specific edge
            assign neighbor_matrix[g] = get_neighbors(g[3:0]);
            // 2. If this edge is currently visited, activate its neighbors
            assign active_neighbors[g] = visited[g] ? neighbor_matrix[g] : 12'd0;
        end
    endgenerate

    always_comb begin
        // Combine all active neighbors (OR-tree) and mask with actual existing edges
        next_visited = visited | (
            active_neighbors[0] | active_neighbors[1] | active_neighbors[2] | 
            active_neighbors[3] | active_neighbors[4] | active_neighbors[5] | 
            active_neighbors[6] | active_neighbors[7] | active_neighbors[8] | 
            active_neighbors[9] | active_neighbors[10]| active_neighbors[11]
        ) & edges_in;
    end

    // --- Sequential Logic ---
    always @(posedge clk) begin
        if (rst) begin
            state <= S_IDLE;
            visited <= 0;
            done <= 0;
            busy <= 0;
            is_connected <= 0;
        end else begin
            case (state)
                S_IDLE: begin
                    done <= 0;
                    if (start) begin
                        if (edges_in == 0) begin
                            is_connected <= 0;
                            done <= 1;
                        end else begin
                            busy <= 1;
                            state <= S_SEED;
                        end
                    end
                end

                S_SEED: begin
                    // "Two's Complement Isolation": Finds the lowest set bit instantly
                    visited <= edges_in & (~edges_in + 1); 
                    state <= S_FLOOD;
                end

                S_FLOOD: begin
                    visited <= next_visited;
                    // If no new nodes were visited this cycle, the flood has stabilized
                    if (visited == next_visited) begin
                        state <= S_CHECK;
                    end
                end

                S_CHECK: begin
                    // If visited == edges_in, we reached every edge -> Connected!
                    is_connected <= (visited == edges_in);
                    done <= 1;
                    busy <= 0;
                    state <= S_IDLE;
                end
            endcase
        end
    end
endmodule // cube_connectivity


module cube_solver (
    input  wire       clk,
    input  wire       rst,
    input  wire       run,
    output reg [11:0] total_found,
    output reg        solver_done
);

    reg  [11:0] current_test_cube;
    reg         conn_start;
    wire        conn_is_connected;
    wire        conn_done;
    
    // 1. Connectivity Instance
    cube_connectivity checker_inst (
        .clk(clk),
        .rst(rst),
        .start(conn_start),
        .edges_in(current_test_cube),
        .is_connected(conn_is_connected),
        .done(conn_done)
    );

    // 2. Canonical Instance (The Uniqueness Filter)
    wire is_canon;
    is_canonical canon_inst (
        .cube_in(current_test_cube),
        .is_canon(is_canon)
    );

    // 3. 3D Logic
    wire has_width = |(current_test_cube & 12'b010100000101);
    wire has_depth = |(current_test_cube & 12'b101000001010);
    wire has_vert  = |(current_test_cube & 12'b000011110000);
    wire is_3d_valid = has_width && has_depth && has_vert;

    // State Machine
    reg [1:0] state;
    localparam S_IDLE = 0, S_TEST = 1, S_WAIT = 2, S_NEXT = 3;

    always @(posedge clk) begin
        if (rst) begin
            state <= S_IDLE;
            current_test_cube <= 0;
            total_found <= 0;
            solver_done <= 0;
            conn_start <= 0;
        end else begin
            case (state)
                S_IDLE: begin
                    if (run) begin
                        state <= S_TEST;
                        current_test_cube <= 1;
                        total_found <= 0;
                        solver_done <= 0;
                    end
                end

                S_TEST: begin
                    conn_start <= 1;
                    state <= S_WAIT;
                end

                S_WAIT: begin
                    conn_start <= 0;
                    if (conn_done) begin
                        // *** THE BIG CHECK ***
                        // Only count if Connected AND 3D AND Canonical
                        if (conn_is_connected && is_3d_valid && is_canon) begin
                             total_found <= total_found + 1;
                        end
                        state <= S_NEXT;
                    end
                end

                S_NEXT: begin
		   // stop before the last one, the fully connected cube
                    if (current_test_cube == 12'hFFE) begin
                        solver_done <= 1;
                        state <= S_IDLE;
                    end else begin
                        current_test_cube <= current_test_cube + 1;
                        state <= S_TEST;
                    end
                end
            endcase // case (state)
        end // else: !if(rst)
    end // always @ (posedge clk)
endmodule // cube_solver

