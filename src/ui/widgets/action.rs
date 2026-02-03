//! Widget actions - what widgets report happened
//!
//! These actions define the interface between widgets and App.

use std::path::PathBuf;

/// Actions that widgets can return from key handling.
/// App dispatches these to update other state.
#[derive(Debug, Clone, PartialEq)]
pub enum Action {
    /// No action, key was handled internally
    None,

    /// Key was not handled, pass to parent
    Ignored,

    // File list actions
    /// File was selected (Enter on file)
    FileSelected(PathBuf),

    // PR list actions
    /// PR was selected
    PrSelected(u64),
    /// Checkout PR
    CheckoutPr(u64),

    // Review actions
    /// Open review modal
    OpenReviewModal(ReviewActionType),
}

/// Types of review actions
#[derive(Debug, Clone, PartialEq)]
pub enum ReviewActionType {
    Approve { pr_number: u64 },
    RequestChanges { pr_number: u64 },
    Comment { pr_number: u64 },
    LineComment { pr_number: u64, path: String, line: u32 },
}
