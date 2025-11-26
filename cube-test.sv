module tb;
    reg clk;
    reg rst;
    reg run;
    wire [11:0] total_found;
    wire solver_done;

    // Instantiate the Solver
    cube_solver dut (
        .clk(clk),
        .rst(rst),
        .run(run),
        .total_found(total_found),
        .solver_done(solver_done)
    );

    // Fast Clock
    always #1 clk = ~clk;

    initial begin
        $dumpfile("dump.vcd"); $dumpvars;
        clk = 0;
        rst = 1;
        run = 0;

        // Reset Pulse
        #10 rst = 0;
        
        // Start the engine
        $display("Starting Hardware Solver...");
        #10 run = 1;
        #2  run = 0;

        // Wait for the "Done" signal
        wait(solver_done);

        $display("------------------------------------------------");
        $display("HARDWARE SOLVER FINISHED");
        $display("Total Valid 3D Connected Cubes: %d", total_found);
        $display("------------------------------------------------");

        $finish;
    end
endmodule
